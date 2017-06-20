// Copyright 2014 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this ***REMOVED***le except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the speci***REMOVED***c language governing permissions and
// limitations under the License.

package local

import (
	"bu***REMOVED***o"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/***REMOVED***lepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/model"

	"github.com/prometheus/prometheus/storage/local/chunk"
	"github.com/prometheus/prometheus/storage/local/codable"
	"github.com/prometheus/prometheus/storage/local/index"
	"github.com/prometheus/prometheus/util/flock"
)

const (
	// Version of the storage as it can be found in the version ***REMOVED***le.
	// Increment to protect against incompatible changes.
	Version         = 1
	versionFileName = "VERSION"

	seriesFileSuf***REMOVED***x     = ".db"
	seriesTempFileSuf***REMOVED***x = ".db.tmp"
	seriesDirNameLen     = 2 // How many bytes of the ***REMOVED***ngerprint in dir name.
	hintFileSuf***REMOVED***x       = ".hint"

	mappingsFileName      = "mappings.db"
	mappingsTempFileName  = "mappings.db.tmp"
	mappingsFormatVersion = 1
	mappingsMagicString   = "PrometheusMappings"

	dirtyFileName = "DIRTY"

	***REMOVED***leBufSize = 1 << 16 // 64kiB.

	chunkHeaderLen             = 17
	chunkHeaderTypeOffset      = 0
	chunkHeaderFirstTimeOffset = 1
	chunkHeaderLastTimeOffset  = 9
	chunkLenWithHeader         = chunk.ChunkLen + chunkHeaderLen
	chunkMaxBatchSize          = 62 // Max no. of chunks to load or write in
	// one batch.  Note that 62 is the largest number of chunks that ***REMOVED***t
	// into 64kiB on disk because chunkHeaderLen is added to each 1k chunk.

	indexingMaxBatchSize  = 1024 * 1024
	indexingBatchTimeout  = 500 * time.Millisecond // Commit batch when idle for that long.
	indexingQueueCapacity = 1024 * 256
)

var fpLen = len(model.Fingerprint(0).String()) // Length of a ***REMOVED***ngerprint as string.

const (
	flagHeadChunkPersisted byte = 1 << iota
	// Add more flags here like:
	// flagFoo
	// flagBar
)

type indexingOpType byte

const (
	add indexingOpType = iota
	remove
)

type indexingOp struct {
	***REMOVED***ngerprint model.Fingerprint
	metric      model.Metric
	opType      indexingOpType
}

// A Persistence is used by a Storage implementation to store samples
// persistently across restarts. The methods are only goroutine-safe if
// explicitly marked as such below. The chunk-related methods persistChunks,
// dropChunks, loadChunks, and loadChunkDescs can be called concurrently with
// each other if each call refers to a different ***REMOVED***ngerprint.
type persistence struct {
	basePath string

	archivedFingerprintToMetrics   *index.FingerprintMetricIndex
	archivedFingerprintToTimeRange *index.FingerprintTimeRangeIndex
	labelPairToFingerprints        *index.LabelPairFingerprintIndex
	labelNameToLabelValues         *index.LabelNameLabelValuesIndex

	indexingQueue   chan indexingOp
	indexingStopped chan struct{}
	indexingFlush   chan chan int

	indexingQueueLength     prometheus.Gauge
	indexingQueueCapacity   prometheus.Metric
	indexingBatchSizes      prometheus.Summary
	indexingBatchDuration   prometheus.Summary
	checkpointDuration      prometheus.Summary
	checkpointLastDuration  prometheus.Gauge
	checkpointLastSize      prometheus.Gauge
	checkpointChunksWritten prometheus.Summary
	dirtyCounter            prometheus.Counter
	startedDirty            prometheus.Gauge
	checkpointing           prometheus.Gauge
	seriesChunksPersisted   prometheus.Histogram

	dirtyMtx       sync.Mutex     // Protects dirty and becameDirty.
	dirty          bool           // true if persistence was started in dirty state.
	becameDirty    bool           // true if an inconsistency came up during runtime.
	pedanticChecks bool           // true if crash recovery should check each series.
	dirtyFileName  string         // The ***REMOVED***le used for locking and to mark dirty state.
	fLock          flock.Releaser // The ***REMOVED***le lock to protect against concurrent usage.

	shouldSync syncStrategy

	minShrinkRatio float64 // How much a series ***REMOVED***le has to shrink to justify dropping chunks.

	bufPool sync.Pool
}

// newPersistence returns a newly allocated persistence backed by local disk storage, ready to use.
func newPersistence(
	basePath string,
	dirty, pedanticChecks bool,
	shouldSync syncStrategy,
	minShrinkRatio float64,
) (*persistence, error) {
	dirtyPath := ***REMOVED***lepath.Join(basePath, dirtyFileName)
	versionPath := ***REMOVED***lepath.Join(basePath, versionFileName)

	if versionData, err := ioutil.ReadFile(versionPath); err == nil {
		if persistedVersion, err := strconv.Atoi(strings.TrimSpace(string(versionData))); err != nil {
			return nil, fmt.Errorf("cannot parse content of %s: %s", versionPath, versionData)
		} ***REMOVED*** if persistedVersion != Version {
			return nil, fmt.Errorf("found storage version %d on disk, need version %d - please wipe storage or run a version of Prometheus compatible with storage version %d", persistedVersion, Version, persistedVersion)
		}
	} ***REMOVED*** if os.IsNotExist(err) {
		// No version ***REMOVED***le found. Let's create the directory (in case
		// it's not there yet) and then check if it is actually
		// empty. If not, we have found an old storage directory without
		// version ***REMOVED***le, so we have to bail out.
		if err := os.MkdirAll(basePath, 0700); err != nil {
			if abspath, e := ***REMOVED***lepath.Abs(basePath); e == nil {
				return nil, fmt.Errorf("cannot create persistent directory %s: %s", abspath, err)
			}
			return nil, fmt.Errorf("cannot create persistent directory %s: %s", basePath, err)
		}
		***REMOVED***s, err := ioutil.ReadDir(basePath)
		if err != nil {
			return nil, err
		}
		***REMOVED***lesPresent := len(***REMOVED***s)
		for i := range ***REMOVED***s {
			switch {
			case ***REMOVED***s[i].Name() == "lost+found" && ***REMOVED***s[i].IsDir():
				***REMOVED***lesPresent--
			case strings.HasPre***REMOVED***x(***REMOVED***s[i].Name(), "."):
				***REMOVED***lesPresent--
			}
		}
		if ***REMOVED***lesPresent > 0 {
			return nil, fmt.Errorf("found existing ***REMOVED***les in storage path that do not look like storage ***REMOVED***les compatible with this version of Prometheus; please delete the ***REMOVED***les in the storage path or choose a different storage path")
		}
		// Finally we can write our own version into a new version ***REMOVED***le.
		***REMOVED***le, err := os.Create(versionPath)
		if err != nil {
			return nil, err
		}
		defer ***REMOVED***le.Close()
		if _, err := fmt.Fprintf(***REMOVED***le, "%d\n", Version); err != nil {
			return nil, err
		}
	} ***REMOVED*** {
		return nil, err
	}

	fLock, dirty***REMOVED***leExisted, err := flock.New(dirtyPath)
	if err != nil {
		log.Errorf("Could not lock %s, Prometheus already running?", dirtyPath)
		return nil, err
	}
	if dirty***REMOVED***leExisted {
		dirty = true
	}

	archivedFingerprintToMetrics, err := index.NewFingerprintMetricIndex(basePath)
	if err != nil {
		// At this point, we could simply blow away the archived
		// ***REMOVED***ngerprint-to-metric index. However, then we would lose
		// _all_ archived metrics. So better give the user an
		// opportunity to repair the LevelDB with a 3rd party tool.
		log.Errorf("Could not open the ***REMOVED***ngerprint-to-metric index for archived series. Please try a 3rd party tool to repair LevelDB in directory %q. If unsuccessful or undesired, delete the whole directory and restart Prometheus for crash recovery. You will lose all archived time series.", ***REMOVED***lepath.Join(basePath, index.FingerprintToMetricDir))
		return nil, err
	}
	archivedFingerprintToTimeRange, err := index.NewFingerprintTimeRangeIndex(basePath)
	if err != nil {
		// We can recover the archived ***REMOVED***ngerprint-to-timerange index,
		// so blow it away and set ourselves dirty. Then re-open the now
		// empty index.
		if err := index.DeleteFingerprintTimeRangeIndex(basePath); err != nil {
			return nil, err
		}
		dirty = true
		if archivedFingerprintToTimeRange, err = index.NewFingerprintTimeRangeIndex(basePath); err != nil {
			return nil, err
		}
	}

	p := &persistence{
		basePath: basePath,

		archivedFingerprintToMetrics:   archivedFingerprintToMetrics,
		archivedFingerprintToTimeRange: archivedFingerprintToTimeRange,

		indexingQueue:   make(chan indexingOp, indexingQueueCapacity),
		indexingStopped: make(chan struct{}),
		indexingFlush:   make(chan chan int),

		indexingQueueLength: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "indexing_queue_length",
			Help:      "The number of metrics waiting to be indexed.",
		}),
		indexingQueueCapacity: prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(namespace, subsystem, "indexing_queue_capacity"),
				"The capacity of the indexing queue.",
				nil, nil,
			),
			prometheus.GaugeValue,
			float64(indexingQueueCapacity),
		),
		indexingBatchSizes: prometheus.NewSummary(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "indexing_batch_sizes",
				Help:      "Quantiles for indexing batch sizes (number of metrics per batch).",
			},
		),
		indexingBatchDuration: prometheus.NewSummary(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "indexing_batch_duration_seconds",
				Help:      "Quantiles for batch indexing duration in seconds.",
			},
		),
		checkpointLastDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "checkpoint_last_duration_seconds",
			Help:      "The duration in seconds it took to last checkpoint open chunks and chunks yet to be persisted.",
		}),
		checkpointDuration: prometheus.NewSummary(prometheus.SummaryOpts{
			Namespace:  namespace,
			Subsystem:  subsystem,
			Objectives: map[float64]float64{},
			Name:       "checkpoint_duration_seconds",
			Help:       "The duration in seconds taken for checkpointing open chunks and chunks yet to be persisted",
		}),
		checkpointLastSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "checkpoint_last_size_bytes",
			Help:      "The size of the last checkpoint of open chunks and chunks yet to be persisted",
		}),
		checkpointChunksWritten: prometheus.NewSummary(prometheus.SummaryOpts{
			Namespace:  namespace,
			Subsystem:  subsystem,
			Objectives: map[float64]float64{},
			Name:       "checkpoint_series_chunks_written",
			Help:       "The number of chunk written per series while checkpointing open chunks and chunks yet to be persisted.",
		}),
		dirtyCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "inconsistencies_total",
			Help:      "A counter incremented each time an inconsistency in the local storage is detected. If this is greater zero, restart the server as soon as possible.",
		}),
		startedDirty: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "started_dirty",
			Help:      "Whether the local storage was found to be dirty (and crash recovery occurred) during Prometheus startup.",
		}),
		checkpointing: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "checkpointing",
			Help:      "1 if the storage is checkpointing, 0 otherwise.",
		}),
		seriesChunksPersisted: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "series_chunks_persisted",
			Help:      "The number of chunks persisted per series.",
			// Even with 4 bytes per sample, you're not going to get more than 85
			// chunks in 6 hours for a time series with 1s resolution.
			Buckets: []float64{1, 2, 4, 8, 16, 32, 64, 128},
		}),
		dirty:          dirty,
		pedanticChecks: pedanticChecks,
		dirtyFileName:  dirtyPath,
		fLock:          fLock,
		shouldSync:     shouldSync,
		minShrinkRatio: minShrinkRatio,
		// Create buffers of length 3*chunkLenWithHeader by default because that is still reasonably small
		// and at the same time enough for many uses. The contract is to never return buffer smaller than
		// that to the pool so that callers can rely on a minimum buffer size.
		bufPool: sync.Pool{New: func() interface{} { return make([]byte, 0, 3*chunkLenWithHeader) }},
	}

	if p.dirty {
		// Blow away the label indexes. We'll rebuild them later.
		if err := index.DeleteLabelPairFingerprintIndex(basePath); err != nil {
			return nil, err
		}
		if err := index.DeleteLabelNameLabelValuesIndex(basePath); err != nil {
			return nil, err
		}
	}
	labelPairToFingerprints, err := index.NewLabelPairFingerprintIndex(basePath)
	if err != nil {
		return nil, err
	}
	labelNameToLabelValues, err := index.NewLabelNameLabelValuesIndex(basePath)
	if err != nil {
		return nil, err
	}
	p.labelPairToFingerprints = labelPairToFingerprints
	p.labelNameToLabelValues = labelNameToLabelValues

	return p, nil
}

func (p *persistence) run() {
	p.processIndexingQueue()
}

// Describe implements prometheus.Collector.
func (p *persistence) Describe(ch chan<- *prometheus.Desc) {
	ch <- p.indexingQueueLength.Desc()
	ch <- p.indexingQueueCapacity.Desc()
	p.indexingBatchSizes.Describe(ch)
	p.indexingBatchDuration.Describe(ch)
	ch <- p.checkpointDuration.Desc()
	ch <- p.checkpointLastDuration.Desc()
	ch <- p.checkpointLastSize.Desc()
	ch <- p.checkpointChunksWritten.Desc()
	ch <- p.checkpointing.Desc()
	ch <- p.dirtyCounter.Desc()
	ch <- p.startedDirty.Desc()
	ch <- p.seriesChunksPersisted.Desc()
}

// Collect implements prometheus.Collector.
func (p *persistence) Collect(ch chan<- prometheus.Metric) {
	p.indexingQueueLength.Set(float64(len(p.indexingQueue)))

	ch <- p.indexingQueueLength
	ch <- p.indexingQueueCapacity
	p.indexingBatchSizes.Collect(ch)
	p.indexingBatchDuration.Collect(ch)
	ch <- p.checkpointDuration
	ch <- p.checkpointLastDuration
	ch <- p.checkpointLastSize
	ch <- p.checkpointChunksWritten
	ch <- p.checkpointing
	ch <- p.dirtyCounter
	ch <- p.startedDirty
	ch <- p.seriesChunksPersisted
}

// isDirty returns the dirty flag in a goroutine-safe way.
func (p *persistence) isDirty() bool {
	p.dirtyMtx.Lock()
	defer p.dirtyMtx.Unlock()
	return p.dirty
}

// setDirty flags the storage as dirty in a goroutine-safe way. The provided
// error will be logged as a reason the ***REMOVED***rst time the storage is flagged as dirty.
func (p *persistence) setDirty(err error) {
	p.dirtyCounter.Inc()
	p.dirtyMtx.Lock()
	defer p.dirtyMtx.Unlock()
	if p.becameDirty {
		return
	}
	p.dirty = true
	p.becameDirty = true
	log.With("error", err).Error("The storage is now inconsistent. Restart Prometheus ASAP to initiate recovery.")
}

// ***REMOVED***ngerprintsForLabelPair returns the ***REMOVED***ngerprints for the given label
// pair. This method is goroutine-safe but take into account that metrics queued
// for indexing with IndexMetric might not have made it into the index
// yet. (Same applies correspondingly to UnindexMetric.)
func (p *persistence) ***REMOVED***ngerprintsForLabelPair(lp model.LabelPair) model.Fingerprints {
	fps, _, err := p.labelPairToFingerprints.Lookup(lp)
	if err != nil {
		p.setDirty(fmt.Errorf("error in method ***REMOVED***ngerprintsForLabelPair(%v): %s", lp, err))
		return nil
	}
	return fps
}

// labelValuesForLabelName returns the label values for the given label
// name. This method is goroutine-safe but take into account that metrics queued
// for indexing with IndexMetric might not have made it into the index
// yet. (Same applies correspondingly to UnindexMetric.)
func (p *persistence) labelValuesForLabelName(ln model.LabelName) (model.LabelValues, error) {
	lvs, _, err := p.labelNameToLabelValues.Lookup(ln)
	if err != nil {
		p.setDirty(fmt.Errorf("error in method labelValuesForLabelName(%v): %s", ln, err))
		return nil, err
	}
	return lvs, nil
}

// persistChunks persists a number of consecutive chunks of a series. It is the
// caller's responsibility to not modify the chunks concurrently and to not
// persist or drop anything for the same ***REMOVED***ngerprint concurrently. It returns
// the (zero-based) index of the ***REMOVED***rst persisted chunk within the series
// ***REMOVED***le. In case of an error, the returned index is -1 (to avoid the
// misconception that the chunk was written at position 0).
//
// Returning an error signals problems with the series ***REMOVED***le. In this case, the
// caller should quarantine the series.
func (p *persistence) persistChunks(fp model.Fingerprint, chunks []chunk.Chunk) (index int, err error) {
	f, err := p.openChunkFileForWriting(fp)
	if err != nil {
		return -1, err
	}
	defer p.closeChunkFile(f)

	if err := p.writeChunks(f, chunks); err != nil {
		return -1, err
	}

	// Determine index within the ***REMOVED***le.
	offset, err := f.Seek(0, os.SEEK_CUR)
	if err != nil {
		return -1, err
	}
	index, err = chunkIndexForOffset(offset)
	if err != nil {
		return -1, err
	}

	return index - len(chunks), err
}

// loadChunks loads a group of chunks of a timeseries by their index. The chunk
// with the earliest time will have index 0, the following ones will have
// incrementally larger indexes. The indexOffset denotes the offset to be added to
// each index in indexes. It is the caller's responsibility to not persist or
// drop anything for the same ***REMOVED***ngerprint concurrently.
func (p *persistence) loadChunks(fp model.Fingerprint, indexes []int, indexOffset int) ([]chunk.Chunk, error) {
	f, err := p.openChunkFileForReading(fp)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	chunks := make([]chunk.Chunk, 0, len(indexes))
	buf := p.bufPool.Get().([]byte)
	defer func() {
		// buf may change below. An unwrapped 'defer p.bufPool.Put(buf)'
		// would only put back the original buf.
		p.bufPool.Put(buf)
	}()

	for i := 0; i < len(indexes); i++ {
		// This loads chunks in batches. A batch is a streak of
		// consecutive chunks, read from disk in one go.
		batchSize := 1
		if _, err := f.Seek(offsetForChunkIndex(indexes[i]+indexOffset), os.SEEK_SET); err != nil {
			return nil, err
		}

		for ; batchSize < chunkMaxBatchSize &&
			i+1 < len(indexes) &&
			indexes[i]+1 == indexes[i+1]; i, batchSize = i+1, batchSize+1 {
		}
		readSize := batchSize * chunkLenWithHeader
		if cap(buf) < readSize {
			buf = make([]byte, readSize)
		}
		buf = buf[:readSize]

		if _, err := io.ReadFull(f, buf); err != nil {
			return nil, err
		}
		for c := 0; c < batchSize; c++ {
			chunk, err := chunk.NewForEncoding(chunk.Encoding(buf[c*chunkLenWithHeader+chunkHeaderTypeOffset]))
			if err != nil {
				return nil, err
			}
			if err := chunk.UnmarshalFromBuf(buf[c*chunkLenWithHeader+chunkHeaderLen:]); err != nil {
				return nil, err
			}
			chunks = append(chunks, chunk)
		}
	}
	chunk.Ops.WithLabelValues(chunk.Load).Add(float64(len(chunks)))
	atomic.AddInt64(&chunk.NumMemChunks, int64(len(chunks)))
	return chunks, nil
}

// loadChunkDescs loads the chunk.Descs for a series from disk. offsetFromEnd is
// the number of chunk.Descs to skip from the end of the series ***REMOVED***le. It is the
// caller's responsibility to not persist or drop anything for the same
// ***REMOVED***ngerprint concurrently.
func (p *persistence) loadChunkDescs(fp model.Fingerprint, offsetFromEnd int) ([]*chunk.Desc, error) {
	f, err := p.openChunkFileForReading(fp)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	***REMOVED***, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if ***REMOVED***.Size()%int64(chunkLenWithHeader) != 0 {
		// The returned error will bubble up and lead to quarantining of the whole series.
		return nil, fmt.Errorf(
			"size of series ***REMOVED***le for ***REMOVED***ngerprint %v is %d, which is not a multiple of the chunk length %d",
			fp, ***REMOVED***.Size(), chunkLenWithHeader,
		)
	}

	numChunks := int(***REMOVED***.Size())/chunkLenWithHeader - offsetFromEnd
	cds := make([]*chunk.Desc, numChunks)
	chunkTimesBuf := make([]byte, 16)
	for i := 0; i < numChunks; i++ {
		_, err := f.Seek(offsetForChunkIndex(i)+chunkHeaderFirstTimeOffset, os.SEEK_SET)
		if err != nil {
			return nil, err
		}

		_, err = io.ReadAtLeast(f, chunkTimesBuf, 16)
		if err != nil {
			return nil, err
		}
		cds[i] = &chunk.Desc{
			ChunkFirstTime: model.Time(binary.LittleEndian.Uint64(chunkTimesBuf)),
			ChunkLastTime:  model.Time(binary.LittleEndian.Uint64(chunkTimesBuf[8:])),
		}
	}
	chunk.DescOps.WithLabelValues(chunk.Load).Add(float64(len(cds)))
	chunk.NumMemDescs.Add(float64(len(cds)))
	return cds, nil
}

// checkpointSeriesMapAndHeads persists the ***REMOVED***ngerprint to memory-series mapping
// and all non persisted chunks. Do not call concurrently with
// loadSeriesMapAndHeads. This method will only write heads format v2, but
// loadSeriesMapAndHeads can also understand v1.
//
// Description of the ***REMOVED***le format (for both, v1 and v2):
//
// (1) Magic string (const headsMagicString).
//
// (2) Varint-encoded format version (const headsFormatVersion).
//
// (3) Number of series in checkpoint as big-endian uint64.
//
// (4) Repeated once per series:
//
// (4.1) A flag byte, see flag constants above. (Present but unused in v2.)
//
// (4.2) The ***REMOVED***ngerprint as big-endian uint64.
//
// (4.3) The metric as de***REMOVED***ned by codable.Metric.
//
// (4.4) The varint-encoded persistWatermark. (Missing in v1.)
//
// (4.5) The modi***REMOVED***cation time of the series ***REMOVED***le as nanoseconds elapsed since
// January 1, 1970 UTC. -1 if the modi***REMOVED***cation time is unknown or no series ***REMOVED***le
// exists yet. (Missing in v1.)
//
// (4.6) The varint-encoded chunkDescsOffset.
//
// (4.6) The varint-encoded savedFirstTime.
//
// (4.7) The varint-encoded number of chunk descriptors.
//
// (4.8) Repeated once per chunk descriptor, oldest to most recent, either
// variant 4.8.1 (if index < persistWatermark) or variant 4.8.2 (if index >=
// persistWatermark). In v1, everything is variant 4.8.1 except for a
// non-persisted head-chunk (determined by the flags).
//
// (4.8.1.1) The varint-encoded ***REMOVED***rst time.
// (4.8.1.2) The varint-encoded last time.
//
// (4.8.2.1) A byte de***REMOVED***ning the chunk type.
// (4.8.2.2) The chunk itself, marshaled with the Marshal() method.
//
// NOTE: Above, varint encoding is used consistently although uvarint would have
// made more sense in many cases. This was simply a glitch while designing the
// format.
func (p *persistence) checkpointSeriesMapAndHeads(
	ctx context.Context, ***REMOVED***ngerprintToSeries *seriesMap, fpLocker ****REMOVED***ngerprintLocker,
) (err error) {
	log.Info("Checkpointing in-memory metrics and chunks...")
	p.checkpointing.Set(1)
	defer p.checkpointing.Set(0)
	begin := time.Now()
	f, err := os.OpenFile(p.headsTempFileName(), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0640)
	if err != nil {
		return err
	}

	defer func() {
		defer os.Remove(p.headsTempFileName()) // Just in case it was left behind.

		if err != nil {
			// If we already had an error, do not bother to sync,
			// just close, ignoring any further error.
			f.Close()
			return
		}
		syncErr := f.Sync()
		closeErr := f.Close()
		err = syncErr
		if err != nil {
			return
		}
		err = closeErr
		if err != nil {
			return
		}
		err = os.Rename(p.headsTempFileName(), p.headsFileName())
		duration := time.Since(begin)
		p.checkpointDuration.Observe(duration.Seconds())
		p.checkpointLastDuration.Set(duration.Seconds())
		log.Infof("Done checkpointing in-memory metrics and chunks in %v.", duration)
	}()

	w := bu***REMOVED***o.NewWriterSize(f, ***REMOVED***leBufSize)

	if _, err = w.WriteString(headsMagicString); err != nil {
		return err
	}
	var numberOfSeriesOffset int
	if numberOfSeriesOffset, err = codable.EncodeVarint(w, headsFormatVersion); err != nil {
		return err
	}
	numberOfSeriesOffset += len(headsMagicString)
	numberOfSeriesInHeader := uint64(***REMOVED***ngerprintToSeries.length())
	// We have to write the number of series as uint64 because we might need
	// to overwrite it later, and a varint might change byte width then.
	if err = codable.EncodeUint64(w, numberOfSeriesInHeader); err != nil {
		return err
	}

	iter := ***REMOVED***ngerprintToSeries.iter()
	defer func() {
		// Consume the iterator in any case to not leak goroutines.
		for range iter {
		}
	}()

	var realNumberOfSeries uint64
	for m := range iter {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		func() { // Wrapped in function to use defer for unlocking the fp.
			fpLocker.Lock(m.fp)
			defer fpLocker.Unlock(m.fp)

			chunksToPersist := len(m.series.chunkDescs) - m.series.persistWatermark
			if len(m.series.chunkDescs) == 0 {
				// This series was completely purged or archived
				// in the meantime. Ignore.
				return
			}
			realNumberOfSeries++

			// Sanity checks.
			if m.series.chunkDescsOffset < 0 && m.series.persistWatermark > 0 {
				panic("encountered unknown chunk desc offset in combination with positive persist watermark")
			}

			// These are the values to save in the normal case.
			var (
				// persistWatermark is zero as we only checkpoint non-persisted chunks.
				persistWatermark int64
				// chunkDescsOffset is shifted by the original persistWatermark for the same reason.
				chunkDescsOffset = int64(m.series.chunkDescsOffset + m.series.persistWatermark)
				numChunkDescs    = int64(chunksToPersist)
			)
			// However, in the special case of a series being fully
			// persisted but still in memory (i.e. not archived), we
			// need to save a "placeholder", for which we use just
			// the chunk desc of the last chunk. Values have to be
			// adjusted accordingly. (The reason for doing it in
			// this weird way is to keep the checkpoint format
			// compatible with older versions.)
			if chunksToPersist == 0 {
				persistWatermark = 1
				chunkDescsOffset-- // Save one chunk desc after all.
				numChunkDescs = 1
			}

			// seriesFlags left empty in v2.
			if err = w.WriteByte(0); err != nil {
				return
			}
			if err = codable.EncodeUint64(w, uint64(m.fp)); err != nil {
				return
			}
			var buf []byte
			buf, err = codable.Metric(m.series.metric).MarshalBinary()
			if err != nil {
				return
			}
			if _, err = w.Write(buf); err != nil {
				return
			}
			if _, err = codable.EncodeVarint(w, persistWatermark); err != nil {
				return
			}
			if m.series.modTime.IsZero() {
				if _, err = codable.EncodeVarint(w, -1); err != nil {
					return
				}
			} ***REMOVED*** {
				if _, err = codable.EncodeVarint(w, m.series.modTime.UnixNano()); err != nil {
					return
				}
			}
			if _, err = codable.EncodeVarint(w, chunkDescsOffset); err != nil {
				return
			}
			if _, err = codable.EncodeVarint(w, int64(m.series.savedFirstTime)); err != nil {
				return
			}
			if _, err = codable.EncodeVarint(w, numChunkDescs); err != nil {
				return
			}
			if chunksToPersist == 0 {
				// Save the one placeholder chunk desc for a fully persisted series.
				chunkDesc := m.series.chunkDescs[len(m.series.chunkDescs)-1]
				if _, err = codable.EncodeVarint(w, int64(chunkDesc.FirstTime())); err != nil {
					return
				}
				lt, err := chunkDesc.LastTime()
				if err != nil {
					return
				}
				if _, err = codable.EncodeVarint(w, int64(lt)); err != nil {
					return
				}
			} ***REMOVED*** {
				// Save (only) the non-persisted chunks.
				for _, chunkDesc := range m.series.chunkDescs[m.series.persistWatermark:] {
					if err = w.WriteByte(byte(chunkDesc.C.Encoding())); err != nil {
						return
					}
					if err = chunkDesc.C.Marshal(w); err != nil {
						return
					}
					p.checkpointChunksWritten.Observe(float64(chunksToPersist))
				}
			}
			// Series is checkpointed now, so declare it clean. In case the entire
			// checkpoint fails later on, this is ***REMOVED***ne, as the storage's series
			// maintenance will mark these series newly dirty again, continuously
			// increasing the total number of dirty series as seen by the storage.
			// This has the effect of triggering a new checkpoint attempt even
			// earlier than if we hadn't incorrectly set "dirty" to "false" here
			// already.
			m.series.dirty = false
		}()
		if err != nil {
			return err
		}
	}
	if err = w.Flush(); err != nil {
		return err
	}
	if realNumberOfSeries != numberOfSeriesInHeader {
		// The number of series has changed in the meantime.
		// Rewrite it in the header.
		if _, err = f.Seek(int64(numberOfSeriesOffset), os.SEEK_SET); err != nil {
			return err
		}
		if err = codable.EncodeUint64(f, realNumberOfSeries); err != nil {
			return err
		}
	}
	info, err := f.Stat()
	if err != nil {
		return err
	}
	p.checkpointLastSize.Set(float64(info.Size()))
	return err
}

// loadSeriesMapAndHeads loads the ***REMOVED***ngerprint to memory-series mapping and all
// the chunks contained in the checkpoint (and thus not yet persisted to series
// ***REMOVED***les). The method is capable of loading the checkpoint format v1 and v2. If
// recoverable corruption is detected, or if the dirty flag was set from the
// beginning, crash recovery is run, which might take a while. If an
// unrecoverable error is encountered, it is returned. Call this method during
// start-up while nothing ***REMOVED*** is running in storage land. This method is
// utterly goroutine-unsafe.
func (p *persistence) loadSeriesMapAndHeads() (sm *seriesMap, chunksToPersist int64, err error) {
	***REMOVED***ngerprintToSeries := make(map[model.Fingerprint]*memorySeries)
	sm = &seriesMap{m: ***REMOVED***ngerprintToSeries}

	defer func() {
		if p.dirty {
			log.Warn("Persistence layer appears dirty.")
			p.startedDirty.Set(1)
			err = p.recoverFromCrash(***REMOVED***ngerprintToSeries)
			if err != nil {
				sm = nil
			}
		} ***REMOVED*** {
			p.startedDirty.Set(0)
		}
	}()

	hs := newHeadsScanner(p.headsFileName())
	defer hs.close()
	for hs.scan() {
		***REMOVED***ngerprintToSeries[hs.fp] = hs.series
	}
	if os.IsNotExist(hs.err) {
		return sm, 0, nil
	}
	if hs.err != nil {
		p.dirty = true
		log.
			With("***REMOVED***le", p.headsFileName()).
			With("error", hs.err).
			Error("Error reading heads ***REMOVED***le.")
		return sm, 0, hs.err
	}
	return sm, hs.chunksToPersistTotal, nil
}

// dropAndPersistChunks deletes all chunks from a series ***REMOVED***le whose last sample
// time is before beforeTime, and then appends the provided chunks, leaving out
// those whose last sample time is before beforeTime. It returns the timestamp
// of the ***REMOVED***rst sample in the oldest chunk _not_ dropped, the chunk offset
// within the series ***REMOVED***le of the ***REMOVED***rst chunk persisted (out of the provided
// chunks, or - if no chunks were provided - the chunk offset where chunks would
// have been persisted, i.e. the end of the ***REMOVED***le), the number of deleted chunks,
// and true if all chunks of the series have been deleted (in which case the
// returned timestamp will be 0 and must be ignored).  It is the caller's
// responsibility to make sure nothing is persisted or loaded for the same
// ***REMOVED***ngerprint concurrently.
//
// Returning an error signals problems with the series ***REMOVED***le. In this case, the
// caller should quarantine the series.
func (p *persistence) dropAndPersistChunks(
	fp model.Fingerprint, beforeTime model.Time, chunks []chunk.Chunk,
) (
	***REMOVED***rstTimeNotDropped model.Time,
	offset int,
	numDropped int,
	allDropped bool,
	err error,
) {
	// Style note: With the many return values, it was decided to use naked
	// returns in this method. They make the method more readable, but
	// please handle with care!
	if len(chunks) > 0 {
		// We have chunks to persist. First check if those are already
		// too old. If that's the case, the chunks in the series ***REMOVED***le
		// are all too old, too.
		i := 0
		for ; i < len(chunks); i++ {
			var lt model.Time
			lt, err = chunks[i].NewIterator().LastTimestamp()
			if err != nil {
				return
			}
			if !lt.Before(beforeTime) {
				break
			}
		}
		if i < len(chunks) {
			***REMOVED***rstTimeNotDropped = chunks[i].FirstTime()
		}
		if i > 0 || ***REMOVED***rstTimeNotDropped.Before(beforeTime) {
			// Series ***REMOVED***le has to go.
			if numDropped, err = p.deleteSeriesFile(fp); err != nil {
				return
			}
			numDropped += i
			if i == len(chunks) {
				allDropped = true
				return
			}
			// Now simply persist what has to be persisted to a new ***REMOVED***le.
			_, err = p.persistChunks(fp, chunks[i:])
			return
		}
	}

	// If we are here, we have to check the series ***REMOVED***le itself.
	f, err := p.openChunkFileForReading(fp)
	if os.IsNotExist(err) {
		// No series ***REMOVED***le. Only need to create new ***REMOVED***le with chunks to
		// persist, if there are any.
		if len(chunks) == 0 {
			allDropped = true
			err = nil // Do not report not-exist err.
			return
		}
		offset, err = p.persistChunks(fp, chunks)
		return
	}
	if err != nil {
		return
	}
	defer f.Close()

	***REMOVED***, err := f.Stat()
	if err != nil {
		return
	}
	chunksInFile := int(***REMOVED***.Size()) / chunkLenWithHeader
	totalChunks := chunksInFile + len(chunks)

	// Calculate chunk index from minShrinkRatio, to skip unnecessary chunk header reading.
	chunkIndexToStartSeek := 0
	if p.minShrinkRatio < 1 {
		chunkIndexToStartSeek = int(math.Floor(float64(totalChunks) * p.minShrinkRatio))
	}
	if chunkIndexToStartSeek >= chunksInFile {
		chunkIndexToStartSeek = chunksInFile - 1
	}
	numDropped = chunkIndexToStartSeek

	headerBuf := make([]byte, chunkHeaderLen)
	// Find the ***REMOVED***rst chunk in the ***REMOVED***le that should be kept.
	for ; ; numDropped++ {
		_, err = f.Seek(offsetForChunkIndex(numDropped), os.SEEK_SET)
		if err != nil {
			return
		}
		_, err = io.ReadFull(f, headerBuf)
		if err == io.EOF {
			// Close the ***REMOVED***le before trying to delete it. This is necessary on Windows
			// (this will cause the defer f.Close to fail, but the error is silently ignored)
			f.Close()
			// We ran into the end of the ***REMOVED***le without ***REMOVED***nding any chunks that should
			// be kept. Remove the whole ***REMOVED***le.
			if numDropped, err = p.deleteSeriesFile(fp); err != nil {
				return
			}
			if len(chunks) == 0 {
				allDropped = true
				return
			}
			offset, err = p.persistChunks(fp, chunks)
			return
		}
		if err != nil {
			return
		}
		lastTime := model.Time(
			binary.LittleEndian.Uint64(headerBuf[chunkHeaderLastTimeOffset:]),
		)
		if !lastTime.Before(beforeTime) {
			break
		}
	}

	// If numDropped isn't incremented, the minShrinkRatio condition isn't satis***REMOVED***ed.
	if numDropped == chunkIndexToStartSeek {
		// Nothing to drop. Just adjust the return values and append the chunks (if any).
		numDropped = 0
		_, err = f.Seek(offsetForChunkIndex(0), os.SEEK_SET)
		if err != nil {
			return
		}
		_, err = io.ReadFull(f, headerBuf)
		if err != nil {
			return
		}
		***REMOVED***rstTimeNotDropped = model.Time(
			binary.LittleEndian.Uint64(headerBuf[chunkHeaderFirstTimeOffset:]),
		)
		if len(chunks) > 0 {
			offset, err = p.persistChunks(fp, chunks)
		} ***REMOVED*** {
			offset = chunksInFile
		}
		return
	}
	// If we are here, we have to drop some chunks for real. So we need to
	// record ***REMOVED***rstTimeNotDropped from the last read header, seek backwards
	// to the beginning of its header, and start copying everything from
	// there into a new ***REMOVED***le. Then append the chunks to the new ***REMOVED***le.
	***REMOVED***rstTimeNotDropped = model.Time(
		binary.LittleEndian.Uint64(headerBuf[chunkHeaderFirstTimeOffset:]),
	)
	chunk.Ops.WithLabelValues(chunk.Drop).Add(float64(numDropped))
	_, err = f.Seek(-chunkHeaderLen, os.SEEK_CUR)
	if err != nil {
		return
	}

	temp, err := os.OpenFile(p.tempFileNameForFingerprint(fp), os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		return
	}
	defer func() {
		// Close the ***REMOVED***le before trying to rename to it. This is necessary on Windows
		// (this will cause the defer f.Close to fail, but the error is silently ignored)
		f.Close()
		p.closeChunkFile(temp)
		if err == nil {
			err = os.Rename(p.tempFileNameForFingerprint(fp), p.***REMOVED***leNameForFingerprint(fp))
		}
	}()

	written, err := io.Copy(temp, f)
	if err != nil {
		return
	}
	offset = int(written / chunkLenWithHeader)

	if len(chunks) > 0 {
		if err = p.writeChunks(temp, chunks); err != nil {
			return
		}
	}
	return
}

// deleteSeriesFile deletes a series ***REMOVED***le belonging to the provided
// ***REMOVED***ngerprint. It returns the number of chunks that were contained in the
// deleted ***REMOVED***le.
func (p *persistence) deleteSeriesFile(fp model.Fingerprint) (int, error) {
	fname := p.***REMOVED***leNameForFingerprint(fp)
	***REMOVED***, err := os.Stat(fname)
	if os.IsNotExist(err) {
		// Great. The ***REMOVED***le is already gone.
		return 0, nil
	}
	if err != nil {
		return -1, err
	}
	numChunks := int(***REMOVED***.Size() / chunkLenWithHeader)
	if err := os.Remove(fname); err != nil {
		return -1, err
	}
	chunk.Ops.WithLabelValues(chunk.Drop).Add(float64(numChunks))
	return numChunks, nil
}

// quarantineSeriesFile moves a series ***REMOVED***le to the orphaned directory. It also
// writes a hint ***REMOVED***le with the provided quarantine reason and, if series is
// non-nil, the string representation of the metric.
func (p *persistence) quarantineSeriesFile(fp model.Fingerprint, quarantineReason error, metric model.Metric) error {
	var (
		oldName     = p.***REMOVED***leNameForFingerprint(fp)
		orphanedDir = ***REMOVED***lepath.Join(p.basePath, "orphaned", ***REMOVED***lepath.Base(***REMOVED***lepath.Dir(oldName)))
		newName     = ***REMOVED***lepath.Join(orphanedDir, ***REMOVED***lepath.Base(oldName))
		hintName    = newName[:len(newName)-len(seriesFileSuf***REMOVED***x)] + hintFileSuf***REMOVED***x
	)

	renameErr := os.MkdirAll(orphanedDir, 0700)
	if renameErr != nil {
		return renameErr
	}
	renameErr = os.Rename(oldName, newName)
	if os.IsNotExist(renameErr) {
		// Source ***REMOVED***le dosn't exist. That's normal.
		renameErr = nil
	}
	// Write hint ***REMOVED***le even if the rename ended in an error. At least try...
	// And ignore errors writing the hint ***REMOVED***le. It's best effort.
	if f, err := os.Create(hintName); err == nil {
		if metric != nil {
			f.WriteString(metric.String() + "\n")
		} ***REMOVED*** {
			f.WriteString("[UNKNOWN METRIC]\n")
		}
		if quarantineReason != nil {
			f.WriteString(quarantineReason.Error() + "\n")
		} ***REMOVED*** {
			f.WriteString("[UNKNOWN REASON]\n")
		}
		f.Close()
	}
	return renameErr
}

// seriesFileModTime returns the modi***REMOVED***cation time of the series ***REMOVED***le belonging
// to the provided ***REMOVED***ngerprint. In case of an error, the zero value of time.Time
// is returned.
func (p *persistence) seriesFileModTime(fp model.Fingerprint) time.Time {
	var modTime time.Time
	if ***REMOVED***, err := os.Stat(p.***REMOVED***leNameForFingerprint(fp)); err == nil {
		return ***REMOVED***.ModTime()
	}
	return modTime
}

// indexMetric queues the given metric for addition to the indexes needed by
// ***REMOVED***ngerprintsForLabelPair, labelValuesForLabelName, and
// ***REMOVED***ngerprintsModi***REMOVED***edBefore.  If the queue is full, this method blocks until
// the metric can be queued.  This method is goroutine-safe.
func (p *persistence) indexMetric(fp model.Fingerprint, m model.Metric) {
	p.indexingQueue <- indexingOp{fp, m, add}
}

// unindexMetric queues references to the given metric for removal from the
// indexes used for ***REMOVED***ngerprintsForLabelPair, labelValuesForLabelName, and
// ***REMOVED***ngerprintsModi***REMOVED***edBefore. The index of ***REMOVED***ngerprints to archived metrics is
// not affected by this removal. (In fact, never call this method for an
// archived metric. To purge an archived metric, call purgeArchivedMetric.)
// If the queue is full, this method blocks until the metric can be queued. This
// method is goroutine-safe.
func (p *persistence) unindexMetric(fp model.Fingerprint, m model.Metric) {
	p.indexingQueue <- indexingOp{fp, m, remove}
}

// waitForIndexing waits until all items in the indexing queue are processed. If
// queue processing is currently on hold (to gather more ops for batching), this
// method will trigger an immediate start of processing. This method is
// goroutine-safe.
func (p *persistence) waitForIndexing() {
	wait := make(chan int)
	for {
		p.indexingFlush <- wait
		if <-wait == 0 {
			break
		}
	}
}

// archiveMetric persists the mapping of the given ***REMOVED***ngerprint to the given
// metric, together with the ***REMOVED***rst and last timestamp of the series belonging to
// the metric. The caller must have locked the ***REMOVED***ngerprint.
func (p *persistence) archiveMetric(
	fp model.Fingerprint, m model.Metric, ***REMOVED***rst, last model.Time,
) {
	if err := p.archivedFingerprintToMetrics.Put(codable.Fingerprint(fp), codable.Metric(m)); err != nil {
		p.setDirty(fmt.Errorf("error in method archiveMetric inserting ***REMOVED***ngerprint %v into FingerprintToMetrics: %s", fp, err))
		return
	}
	if err := p.archivedFingerprintToTimeRange.Put(codable.Fingerprint(fp), codable.TimeRange{First: ***REMOVED***rst, Last: last}); err != nil {
		p.setDirty(fmt.Errorf("error in method archiveMetric inserting ***REMOVED***ngerprint %v into FingerprintToTimeRange: %s", fp, err))
	}
}

// hasArchivedMetric returns whether the archived metric for the given
// ***REMOVED***ngerprint exists and if yes, what the ***REMOVED***rst and last timestamp in the
// corresponding series is. This method is goroutine-safe.
func (p *persistence) hasArchivedMetric(fp model.Fingerprint) (
	hasMetric bool, ***REMOVED***rstTime, lastTime model.Time,
) {
	***REMOVED***rstTime, lastTime, hasMetric, err := p.archivedFingerprintToTimeRange.Lookup(fp)
	if err != nil {
		p.setDirty(fmt.Errorf("error in method hasArchivedMetric(%v): %s", fp, err))
		hasMetric = false
	}
	return hasMetric, ***REMOVED***rstTime, lastTime
}

// updateArchivedTimeRange updates an archived time range. The caller must make
// sure that the ***REMOVED***ngerprint is currently archived (the time range will
// otherwise be added without the corresponding metric in the archive).
func (p *persistence) updateArchivedTimeRange(
	fp model.Fingerprint, ***REMOVED***rst, last model.Time,
) error {
	return p.archivedFingerprintToTimeRange.Put(codable.Fingerprint(fp), codable.TimeRange{First: ***REMOVED***rst, Last: last})
}

// ***REMOVED***ngerprintsModi***REMOVED***edBefore returns the ***REMOVED***ngerprints of archived timeseries
// that have live samples before the provided timestamp. This method is
// goroutine-safe.
func (p *persistence) ***REMOVED***ngerprintsModi***REMOVED***edBefore(beforeTime model.Time) ([]model.Fingerprint, error) {
	var fp codable.Fingerprint
	var tr codable.TimeRange
	fps := []model.Fingerprint{}
	err := p.archivedFingerprintToTimeRange.ForEach(func(kv index.KeyValueAccessor) error {
		if err := kv.Value(&tr); err != nil {
			return err
		}
		if tr.First.Before(beforeTime) {
			if err := kv.Key(&fp); err != nil {
				return err
			}
			fps = append(fps, model.Fingerprint(fp))
		}
		return nil
	})
	return fps, err
}

// archivedMetric retrieves the archived metric with the given ***REMOVED***ngerprint. This
// method is goroutine-safe.
func (p *persistence) archivedMetric(fp model.Fingerprint) (model.Metric, error) {
	metric, _, err := p.archivedFingerprintToMetrics.Lookup(fp)
	if err != nil {
		p.setDirty(fmt.Errorf("error in method archivedMetric(%v): %s", fp, err))
		return nil, err
	}
	return metric, nil
}

// purgeArchivedMetric deletes an archived ***REMOVED***ngerprint and its corresponding
// metric entirely. It also queues the metric for un-indexing (no need to call
// unindexMetric for the deleted metric.) It does not touch the series ***REMOVED***le,
// though. The caller must have locked the ***REMOVED***ngerprint.
func (p *persistence) purgeArchivedMetric(fp model.Fingerprint) (err error) {
	defer func() {
		if err != nil {
			p.setDirty(fmt.Errorf("error in method purgeArchivedMetric(%v): %s", fp, err))
		}
	}()

	metric, err := p.archivedMetric(fp)
	if err != nil || metric == nil {
		return err
	}
	deleted, err := p.archivedFingerprintToMetrics.Delete(codable.Fingerprint(fp))
	if err != nil {
		return err
	}
	if !deleted {
		log.Errorf("Tried to delete non-archived ***REMOVED***ngerprint %s from archivedFingerprintToMetrics index. This should never happen.", fp)
	}
	deleted, err = p.archivedFingerprintToTimeRange.Delete(codable.Fingerprint(fp))
	if err != nil {
		return err
	}
	if !deleted {
		log.Errorf("Tried to delete non-archived ***REMOVED***ngerprint %s from archivedFingerprintToTimeRange index. This should never happen.", fp)
	}
	p.unindexMetric(fp, metric)
	return nil
}

// unarchiveMetric deletes an archived ***REMOVED***ngerprint and its metric, but (in
// contrast to purgeArchivedMetric) does not un-index the metric.  If a metric
// was actually deleted, the method returns true and the ***REMOVED***rst time and last
// time of the deleted metric. The caller must have locked the ***REMOVED***ngerprint.
func (p *persistence) unarchiveMetric(fp model.Fingerprint) (deletedAnything bool, err error) {
	// An error returned here will bubble up and lead to quarantining of the
	// series, so no setDirty required.
	deleted, err := p.archivedFingerprintToMetrics.Delete(codable.Fingerprint(fp))
	if err != nil || !deleted {
		return false, err
	}
	deleted, err = p.archivedFingerprintToTimeRange.Delete(codable.Fingerprint(fp))
	if err != nil {
		return false, err
	}
	if !deleted {
		log.Errorf("Tried to delete non-archived ***REMOVED***ngerprint %s from archivedFingerprintToTimeRange index. This should never happen.", fp)
	}
	return true, nil
}

// close flushes the indexing queue and other buffered data and releases any
// held resources. It also removes the dirty marker ***REMOVED***le if successful and if
// the persistence is currently not marked as dirty.
func (p *persistence) close() error {
	close(p.indexingQueue)
	<-p.indexingStopped

	var lastError, dirtyFileRemoveError error
	if err := p.archivedFingerprintToMetrics.Close(); err != nil {
		lastError = err
		log.Error("Error closing archivedFingerprintToMetric index DB: ", err)
	}
	if err := p.archivedFingerprintToTimeRange.Close(); err != nil {
		lastError = err
		log.Error("Error closing archivedFingerprintToTimeRange index DB: ", err)
	}
	if err := p.labelPairToFingerprints.Close(); err != nil {
		lastError = err
		log.Error("Error closing labelPairToFingerprints index DB: ", err)
	}
	if err := p.labelNameToLabelValues.Close(); err != nil {
		lastError = err
		log.Error("Error closing labelNameToLabelValues index DB: ", err)
	}
	if lastError == nil && !p.isDirty() {
		dirtyFileRemoveError = os.Remove(p.dirtyFileName)
	}
	if err := p.fLock.Release(); err != nil {
		lastError = err
		log.Error("Error releasing ***REMOVED***le lock: ", err)
	}
	if dirtyFileRemoveError != nil {
		// On Windows, removing the dirty ***REMOVED***le before unlocking is not
		// possible.  So remove it here if it failed above.
		lastError = os.Remove(p.dirtyFileName)
	}
	return lastError
}

func (p *persistence) dirNameForFingerprint(fp model.Fingerprint) string {
	fpStr := fp.String()
	return ***REMOVED***lepath.Join(p.basePath, fpStr[0:seriesDirNameLen])
}

func (p *persistence) ***REMOVED***leNameForFingerprint(fp model.Fingerprint) string {
	fpStr := fp.String()
	return ***REMOVED***lepath.Join(p.basePath, fpStr[0:seriesDirNameLen], fpStr[seriesDirNameLen:]+seriesFileSuf***REMOVED***x)
}

func (p *persistence) tempFileNameForFingerprint(fp model.Fingerprint) string {
	fpStr := fp.String()
	return ***REMOVED***lepath.Join(p.basePath, fpStr[0:seriesDirNameLen], fpStr[seriesDirNameLen:]+seriesTempFileSuf***REMOVED***x)
}

func (p *persistence) openChunkFileForWriting(fp model.Fingerprint) (*os.File, error) {
	if err := os.MkdirAll(p.dirNameForFingerprint(fp), 0700); err != nil {
		return nil, err
	}
	return os.OpenFile(p.***REMOVED***leNameForFingerprint(fp), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0640)
	// NOTE: Although the ***REMOVED***le was opened for append,
	//     f.Seek(0, os.SEEK_CUR)
	// would now return '0, nil', so we cannot check for a consistent ***REMOVED***le length right now.
	// However, the chunkIndexForOffset function is doing that check, so a wrong ***REMOVED***le length
	// would still be detected.
}

// closeChunkFile ***REMOVED***rst syncs the provided ***REMOVED***le if mandated so by the sync
// strategy. Then it closes the ***REMOVED***le. Errors are logged.
func (p *persistence) closeChunkFile(f *os.File) {
	if p.shouldSync() {
		if err := f.Sync(); err != nil {
			log.Error("Error syncing ***REMOVED***le:", err)
		}
	}
	if err := f.Close(); err != nil {
		log.Error("Error closing chunk ***REMOVED***le:", err)
	}
}

func (p *persistence) openChunkFileForReading(fp model.Fingerprint) (*os.File, error) {
	return os.Open(p.***REMOVED***leNameForFingerprint(fp))
}

func (p *persistence) headsFileName() string {
	return ***REMOVED***lepath.Join(p.basePath, headsFileName)
}

func (p *persistence) headsTempFileName() string {
	return ***REMOVED***lepath.Join(p.basePath, headsTempFileName)
}

func (p *persistence) mappingsFileName() string {
	return ***REMOVED***lepath.Join(p.basePath, mappingsFileName)
}

func (p *persistence) mappingsTempFileName() string {
	return ***REMOVED***lepath.Join(p.basePath, mappingsTempFileName)
}

func (p *persistence) processIndexingQueue() {
	batchSize := 0
	nameToValues := index.LabelNameLabelValuesMapping{}
	pairToFPs := index.LabelPairFingerprintsMapping{}
	batchTimeout := time.NewTimer(indexingBatchTimeout)
	defer batchTimeout.Stop()

	commitBatch := func() {
		p.indexingBatchSizes.Observe(float64(batchSize))
		defer func(begin time.Time) {
			p.indexingBatchDuration.Observe(time.Since(begin).Seconds())
		}(time.Now())

		if err := p.labelPairToFingerprints.IndexBatch(pairToFPs); err != nil {
			log.Error("Error indexing label pair to ***REMOVED***ngerprints batch: ", err)
			p.setDirty(err)
		}
		if err := p.labelNameToLabelValues.IndexBatch(nameToValues); err != nil {
			log.Error("Error indexing label name to label values batch: ", err)
			p.setDirty(err)
		}
		batchSize = 0
		nameToValues = index.LabelNameLabelValuesMapping{}
		pairToFPs = index.LabelPairFingerprintsMapping{}
		batchTimeout.Reset(indexingBatchTimeout)
	}

	var flush chan chan int
loop:
	for {
		// Only process flush requests if the queue is currently empty.
		if len(p.indexingQueue) == 0 {
			flush = p.indexingFlush
		} ***REMOVED*** {
			flush = nil
		}
		select {
		case <-batchTimeout.C:
			// Only commit if we have something to commit _and_
			// nothing is waiting in the queue to be picked up. That
			// prevents a death spiral if the LookupSet calls below
			// are slow for some reason.
			if batchSize > 0 && len(p.indexingQueue) == 0 {
				commitBatch()
			} ***REMOVED*** {
				batchTimeout.Reset(indexingBatchTimeout)
			}
		case r := <-flush:
			if batchSize > 0 {
				commitBatch()
			}
			r <- len(p.indexingQueue)
		case op, ok := <-p.indexingQueue:
			if !ok {
				if batchSize > 0 {
					commitBatch()
				}
				break loop
			}

			batchSize++
			for ln, lv := range op.metric {
				lp := model.LabelPair{Name: ln, Value: lv}
				baseFPs, ok := pairToFPs[lp]
				if !ok {
					var err error
					baseFPs, _, err = p.labelPairToFingerprints.LookupSet(lp)
					if err != nil {
						log.Errorf("Error looking up label pair %v: %s", lp, err)
						continue
					}
					pairToFPs[lp] = baseFPs
				}
				baseValues, ok := nameToValues[ln]
				if !ok {
					var err error
					baseValues, _, err = p.labelNameToLabelValues.LookupSet(ln)
					if err != nil {
						log.Errorf("Error looking up label name %v: %s", ln, err)
						continue
					}
					nameToValues[ln] = baseValues
				}
				switch op.opType {
				case add:
					baseFPs[op.***REMOVED***ngerprint] = struct{}{}
					baseValues[lv] = struct{}{}
				case remove:
					delete(baseFPs, op.***REMOVED***ngerprint)
					if len(baseFPs) == 0 {
						delete(baseValues, lv)
					}
				default:
					panic("unknown op type")
				}
			}

			if batchSize >= indexingMaxBatchSize {
				commitBatch()
			}
		}
	}
	close(p.indexingStopped)
}

// checkpointFPMappings persists the ***REMOVED***ngerprint mappings. The caller has to
// ensure that the provided mappings are not changed concurrently. This method
// is only called upon shutdown or during crash recovery, when no samples are
// ingested.
//
// Description of the ***REMOVED***le format, v1:
//
// (1) Magic string (const mappingsMagicString).
//
// (2) Uvarint-encoded format version (const mappingsFormatVersion).
//
// (3) Uvarint-encoded number of mappings in fpMappings.
//
// (4) Repeated once per mapping:
//
// (4.1) The raw ***REMOVED***ngerprint as big-endian uint64.
//
// (4.2) The uvarint-encoded number of sub-mappings for the raw ***REMOVED***ngerprint.
//
// (4.3) Repeated once per sub-mapping:
//
// (4.3.1) The uvarint-encoded length of the unique metric string.
// (4.3.2) The unique metric string.
// (4.3.3) The mapped ***REMOVED***ngerprint as big-endian uint64.
func (p *persistence) checkpointFPMappings(fpm fpMappings) (err error) {
	log.Info("Checkpointing ***REMOVED***ngerprint mappings...")
	begin := time.Now()
	f, err := os.OpenFile(p.mappingsTempFileName(), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0640)
	if err != nil {
		return
	}

	defer func() {
		syncErr := f.Sync()
		closeErr := f.Close()
		if err != nil {
			return
		}
		err = syncErr
		if err != nil {
			return
		}
		err = closeErr
		if err != nil {
			return
		}
		err = os.Rename(p.mappingsTempFileName(), p.mappingsFileName())
		duration := time.Since(begin)
		log.Infof("Done checkpointing ***REMOVED***ngerprint mappings in %v.", duration)
	}()

	w := bu***REMOVED***o.NewWriterSize(f, ***REMOVED***leBufSize)

	if _, err = w.WriteString(mappingsMagicString); err != nil {
		return
	}
	if _, err = codable.EncodeUvarint(w, mappingsFormatVersion); err != nil {
		return
	}
	if _, err = codable.EncodeUvarint(w, uint64(len(fpm))); err != nil {
		return
	}

	for fp, mappings := range fpm {
		if err = codable.EncodeUint64(w, uint64(fp)); err != nil {
			return
		}
		if _, err = codable.EncodeUvarint(w, uint64(len(mappings))); err != nil {
			return
		}
		for ms, mappedFP := range mappings {
			if _, err = codable.EncodeUvarint(w, uint64(len(ms))); err != nil {
				return
			}
			if _, err = w.WriteString(ms); err != nil {
				return
			}
			if err = codable.EncodeUint64(w, uint64(mappedFP)); err != nil {
				return
			}
		}
	}
	err = w.Flush()
	return
}

// loadFPMappings loads the ***REMOVED***ngerprint mappings. It also returns the highest
// mapped ***REMOVED***ngerprint and any error encountered. If p.mappingsFileName is not
// found, the method returns (fpMappings{}, 0, nil). Do not call concurrently
// with checkpointFPMappings.
func (p *persistence) loadFPMappings() (fpMappings, model.Fingerprint, error) {
	fpm := fpMappings{}
	var highestMappedFP model.Fingerprint

	f, err := os.Open(p.mappingsFileName())
	if os.IsNotExist(err) {
		return fpm, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	r := bu***REMOVED***o.NewReaderSize(f, ***REMOVED***leBufSize)

	buf := make([]byte, len(mappingsMagicString))
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, 0, err
	}
	magic := string(buf)
	if magic != mappingsMagicString {
		return nil, 0, fmt.Errorf(
			"unexpected magic string, want %q, got %q",
			mappingsMagicString, magic,
		)
	}
	version, err := binary.ReadUvarint(r)
	if version != mappingsFormatVersion || err != nil {
		return nil, 0, fmt.Errorf("unknown ***REMOVED***ngerprint mappings format version, want %d", mappingsFormatVersion)
	}
	numRawFPs, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, 0, err
	}
	for ; numRawFPs > 0; numRawFPs-- {
		rawFP, err := codable.DecodeUint64(r)
		if err != nil {
			return nil, 0, err
		}
		numMappings, err := binary.ReadUvarint(r)
		if err != nil {
			return nil, 0, err
		}
		mappings := make(map[string]model.Fingerprint, numMappings)
		for ; numMappings > 0; numMappings-- {
			lenMS, err := binary.ReadUvarint(r)
			if err != nil {
				return nil, 0, err
			}
			buf := make([]byte, lenMS)
			if _, err := io.ReadFull(r, buf); err != nil {
				return nil, 0, err
			}
			fp, err := codable.DecodeUint64(r)
			if err != nil {
				return nil, 0, err
			}
			mappedFP := model.Fingerprint(fp)
			if mappedFP > highestMappedFP {
				highestMappedFP = mappedFP
			}
			mappings[string(buf)] = mappedFP
		}
		fpm[model.Fingerprint(rawFP)] = mappings
	}
	return fpm, highestMappedFP, nil
}

func (p *persistence) writeChunks(w io.Writer, chunks []chunk.Chunk) error {
	b := p.bufPool.Get().([]byte)
	defer func() {
		// buf may change below. An unwrapped 'defer p.bufPool.Put(buf)'
		// would only put back the original buf.
		p.bufPool.Put(b)
	}()
	numChunks := len(chunks)

	for batchSize := chunkMaxBatchSize; len(chunks) > 0; chunks = chunks[batchSize:] {
		if batchSize > len(chunks) {
			batchSize = len(chunks)
		}
		writeSize := batchSize * chunkLenWithHeader
		if cap(b) < writeSize {
			b = make([]byte, writeSize)
		}
		b = b[:writeSize]

		for i, chunk := range chunks[:batchSize] {
			if err := writeChunkHeader(b[i*chunkLenWithHeader:], chunk); err != nil {
				return err
			}
			if err := chunk.MarshalToBuf(b[i*chunkLenWithHeader+chunkHeaderLen:]); err != nil {
				return err
			}
		}
		if _, err := w.Write(b); err != nil {
			return err
		}
	}
	p.seriesChunksPersisted.Observe(float64(numChunks))
	return nil
}

func offsetForChunkIndex(i int) int64 {
	return int64(i * chunkLenWithHeader)
}

func chunkIndexForOffset(offset int64) (int, error) {
	if int(offset)%chunkLenWithHeader != 0 {
		return -1, fmt.Errorf(
			"offset %d is not a multiple of on-disk chunk length %d",
			offset, chunkLenWithHeader,
		)
	}
	return int(offset) / chunkLenWithHeader, nil
}

func writeChunkHeader(header []byte, c chunk.Chunk) error {
	header[chunkHeaderTypeOffset] = byte(c.Encoding())
	binary.LittleEndian.PutUint64(
		header[chunkHeaderFirstTimeOffset:],
		uint64(c.FirstTime()),
	)
	lt, err := c.NewIterator().LastTimestamp()
	if err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(
		header[chunkHeaderLastTimeOffset:],
		uint64(lt),
	)
	return nil
}
