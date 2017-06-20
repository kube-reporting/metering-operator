// Copyright 2015 The Prometheus Authors
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
	"errors"
	"fmt"
	"io"
	"os"
	"path/***REMOVED***lepath"
	"strings"
	"sync/atomic"

	"github.com/prometheus/common/log"
	"github.com/prometheus/common/model"

	"github.com/prometheus/prometheus/storage/local/chunk"
	"github.com/prometheus/prometheus/storage/local/codable"
	"github.com/prometheus/prometheus/storage/local/index"
)

// recoverFromCrash is called by loadSeriesMapAndHeads if the persistence
// appears to be dirty after the loading (either because the loading resulted in
// an error or because the persistence was dirty from the start). Not goroutine
// safe. Only call before anything ***REMOVED*** is running (except index processing
// queue as started by newPersistence).
func (p *persistence) recoverFromCrash(***REMOVED***ngerprintToSeries map[model.Fingerprint]*memorySeries) error {
	// TODO(beorn): We need proper tests for the crash recovery.
	log.Warn("Starting crash recovery. Prometheus is inoperational until complete.")
	log.Warn("To avoid crash recovery in the future, shut down Prometheus with SIGTERM or a HTTP POST to /-/quit.")

	fpsSeen := map[model.Fingerprint]struct{}{}
	count := 0
	seriesDirNameFmt := fmt.Sprintf("%%0%dx", seriesDirNameLen)

	// Delete the ***REMOVED***ngerprint mapping ***REMOVED***le as it might be stale or
	// corrupt. We'll rebuild the mappings as we go.
	if err := os.RemoveAll(p.mappingsFileName()); err != nil {
		return fmt.Errorf("couldn't remove old ***REMOVED***ngerprint mapping ***REMOVED***le %s: %s", p.mappingsFileName(), err)
	}
	// The mappings to rebuild.
	fpm := fpMappings{}

	log.Info("Scanning ***REMOVED***les.")
	for i := 0; i < 1<<(seriesDirNameLen*4); i++ {
		dirname := ***REMOVED***lepath.Join(p.basePath, fmt.Sprintf(seriesDirNameFmt, i))
		dir, err := os.Open(dirname)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		for ***REMOVED***s := []os.FileInfo{}; err != io.EOF; ***REMOVED***s, err = dir.Readdir(1024) {
			if err != nil {
				dir.Close()
				return err
			}
			for _, ***REMOVED*** := range ***REMOVED***s {
				fp, ok := p.sanitizeSeries(dirname, ***REMOVED***, ***REMOVED***ngerprintToSeries, fpm)
				if ok {
					fpsSeen[fp] = struct{}{}
				}
				count++
				if count%10000 == 0 {
					log.Infof("%d ***REMOVED***les scanned.", count)
				}
			}
		}
		dir.Close()
	}
	log.Infof("File scan complete. %d series found.", len(fpsSeen))

	log.Info("Checking for series without series ***REMOVED***le.")
	for fp, s := range ***REMOVED***ngerprintToSeries {
		if _, seen := fpsSeen[fp]; !seen {
			// fp exists in ***REMOVED***ngerprintToSeries, but has no representation on disk.
			if s.persistWatermark >= len(s.chunkDescs) {
				// Oops, everything including the head chunk was
				// already persisted, but nothing on disk. Or
				// the persistWatermark is plainly wrong. Thus,
				// we lost that series completely. Clean up the
				// remnants.
				delete(***REMOVED***ngerprintToSeries, fp)
				if err := p.purgeArchivedMetric(fp); err != nil {
					// Purging the archived metric didn't work, so try
					// to unindex it, just in case it's in the indexes.
					p.unindexMetric(fp, s.metric)
				}
				log.Warnf("Lost series detected: ***REMOVED***ngerprint %v, metric %v.", fp, s.metric)
				continue
			}
			// If we are here, the only chunks we have are the chunks in the checkpoint.
			// Adjust things accordingly.
			if s.persistWatermark > 0 || s.chunkDescsOffset != 0 {
				minLostChunks := s.persistWatermark + s.chunkDescsOffset
				if minLostChunks <= 0 {
					log.Warnf(
						"Possible loss of chunks for ***REMOVED***ngerprint %v, metric %v.",
						fp, s.metric,
					)
				} ***REMOVED*** {
					log.Warnf(
						"Lost at least %d chunks for ***REMOVED***ngerprint %v, metric %v.",
						minLostChunks, fp, s.metric,
					)
				}
				s.chunkDescs = append(
					make([]*chunk.Desc, 0, len(s.chunkDescs)-s.persistWatermark),
					s.chunkDescs[s.persistWatermark:]...,
				)
				chunk.NumMemDescs.Sub(float64(s.persistWatermark))
				s.persistWatermark = 0
				s.chunkDescsOffset = 0
			}
			maybeAddMapping(fp, s.metric, fpm)
			fpsSeen[fp] = struct{}{} // Add so that fpsSeen is complete.
		}
	}
	log.Info("Check for series without series ***REMOVED***le complete.")

	if err := p.cleanUpArchiveIndexes(***REMOVED***ngerprintToSeries, fpsSeen, fpm); err != nil {
		return err
	}
	if err := p.rebuildLabelIndexes(***REMOVED***ngerprintToSeries); err != nil {
		return err
	}
	// Finally rewrite the mappings ***REMOVED***le if there are any mappings.
	if len(fpm) > 0 {
		if err := p.checkpointFPMappings(fpm); err != nil {
			return err
		}
	}

	p.dirtyMtx.Lock()
	// Only declare storage clean if it didn't become dirty during crash recovery.
	if !p.becameDirty {
		p.dirty = false
	}
	p.dirtyMtx.Unlock()

	log.Warn("Crash recovery complete.")
	return nil
}

// sanitizeSeries sanitizes a series based on its series ***REMOVED***le as de***REMOVED***ned by the
// provided directory and FileInfo.  The method returns the ***REMOVED***ngerprint as
// derived from the directory and ***REMOVED***le name, and whether the provided ***REMOVED***le has
// been sanitized. A ***REMOVED***le that failed to be sanitized is moved into the
// "orphaned" sub-directory, if possible.
//
// The following steps are performed:
//
// - A ***REMOVED***le whose name doesn't comply with the naming scheme of a series ***REMOVED***le is
//   simply moved into the orphaned directory.
//
// - If the size of the series ***REMOVED***le isn't a multiple of the chunk size,
//   extraneous bytes are truncated.  If the truncation fails, the ***REMOVED***le is
//   moved into the orphaned directory.
//
// - A ***REMOVED***le that is empty (after truncation) is deleted.
//
// - A series that is not archived (i.e. it is in the ***REMOVED***ngerprintToSeries map)
//   is checked for consistency of its various parameters (like persist
//   watermark, offset of chunkDescs etc.). In particular, overlap between an
//   in-memory head chunk with the most recent persisted chunk is
//   checked. Inconsistencies are recti***REMOVED***ed.
//
// - A series that is archived (i.e. it is not in the ***REMOVED***ngerprintToSeries map)
//   is checked for its presence in the index of archived series. If it cannot
//   be found there, it is moved into the orphaned directory.
func (p *persistence) sanitizeSeries(
	dirname string, ***REMOVED*** os.FileInfo,
	***REMOVED***ngerprintToSeries map[model.Fingerprint]*memorySeries,
	fpm fpMappings,
) (model.Fingerprint, bool) {
	var (
		fp       model.Fingerprint
		err      error
		***REMOVED***lename = ***REMOVED***lepath.Join(dirname, ***REMOVED***.Name())
		s        *memorySeries
	)

	purge := func() {
		if fp != 0 {
			var metric model.Metric
			if s != nil {
				metric = s.metric
			}
			if err = p.quarantineSeriesFile(
				fp, errors.New("purge during crash recovery"), metric,
			); err == nil {
				return
			}
			log.
				With("***REMOVED***le", ***REMOVED***lename).
				With("error", err).
				Error("Failed to move lost series ***REMOVED***le to orphaned directory.")
		}
		// If we are here, we are either purging an incorrectly named
		// ***REMOVED***le, or quarantining has failed. So simply delete the ***REMOVED***le.
		if err = os.Remove(***REMOVED***lename); err != nil {
			log.
				With("***REMOVED***le", ***REMOVED***lename).
				With("error", err).
				Error("Failed to delete lost series ***REMOVED***le.")
		}
	}

	if len(***REMOVED***.Name()) != fpLen-seriesDirNameLen+len(seriesFileSuf***REMOVED***x) ||
		!strings.HasSuf***REMOVED***x(***REMOVED***.Name(), seriesFileSuf***REMOVED***x) {
		log.Warnf("Unexpected series ***REMOVED***le name %s.", ***REMOVED***lename)
		purge()
		return fp, false
	}
	if fp, err = model.FingerprintFromString(***REMOVED***lepath.Base(dirname) + ***REMOVED***.Name()[:fpLen-seriesDirNameLen]); err != nil {
		log.Warnf("Error parsing ***REMOVED***le name %s: %s", ***REMOVED***lename, err)
		purge()
		return fp, false
	}

	bytesToTrim := ***REMOVED***.Size() % int64(chunkLenWithHeader)
	chunksInFile := int(***REMOVED***.Size()) / chunkLenWithHeader
	modTime := ***REMOVED***.ModTime()
	if bytesToTrim != 0 {
		log.Warnf(
			"Truncating ***REMOVED***le %s to exactly %d chunks, trimming %d extraneous bytes.",
			***REMOVED***lename, chunksInFile, bytesToTrim,
		)
		f, err := os.OpenFile(***REMOVED***lename, os.O_WRONLY, 0640)
		if err != nil {
			log.Errorf("Could not open ***REMOVED***le %s: %s", ***REMOVED***lename, err)
			purge()
			return fp, false
		}
		if err := f.Truncate(***REMOVED***.Size() - bytesToTrim); err != nil {
			log.Errorf("Failed to truncate ***REMOVED***le %s: %s", ***REMOVED***lename, err)
			purge()
			return fp, false
		}
	}
	if chunksInFile == 0 {
		log.Warnf("No chunks left in ***REMOVED***le %s.", ***REMOVED***lename)
		purge()
		return fp, false
	}

	s, ok := ***REMOVED***ngerprintToSeries[fp]
	if ok { // This series is supposed to not be archived.
		if s == nil {
			panic("***REMOVED***ngerprint mapped to nil pointer")
		}
		maybeAddMapping(fp, s.metric, fpm)
		if !p.pedanticChecks &&
			bytesToTrim == 0 &&
			s.chunkDescsOffset != -1 &&
			chunksInFile == s.chunkDescsOffset+s.persistWatermark &&
			modTime.Equal(s.modTime) {
			// Everything is consistent. We are good.
			return fp, true
		}
		// If we are here, we cannot be sure the series ***REMOVED***le is
		// consistent with the checkpoint, so we have to take a closer
		// look.
		if s.headChunkClosed {
			// This is the easy case as we have all chunks on
			// disk. Treat this series as a freshly unarchived one
			// by loading the chunkDescs and setting all parameters
			// based on the loaded chunkDescs.
			cds, err := p.loadChunkDescs(fp, 0)
			if err != nil {
				log.Errorf(
					"Failed to load chunk descriptors for metric %v, ***REMOVED***ngerprint %v: %s",
					s.metric, fp, err,
				)
				purge()
				return fp, false
			}
			log.Warnf(
				"Treating recovered metric %v, ***REMOVED***ngerprint %v, as freshly unarchived, with %d chunks in series ***REMOVED***le.",
				s.metric, fp, len(cds),
			)
			s.chunkDescs = cds
			s.chunkDescsOffset = 0
			s.savedFirstTime = cds[0].FirstTime()
			s.lastTime, err = cds[len(cds)-1].LastTime()
			if err != nil {
				log.Errorf(
					"Failed to determine time of the last sample for metric %v, ***REMOVED***ngerprint %v: %s",
					s.metric, fp, err,
				)
				purge()
				return fp, false
			}
			s.persistWatermark = len(cds)
			s.modTime = modTime
			// Finally, evict again all chunk.Descs except the latest one to save memory.
			s.evictChunkDescs(len(cds) - 1)
			return fp, true
		}
		// This is the tricky one: We have chunks from heads.db, but
		// some of those chunks might already be in the series
		// ***REMOVED***le. Strategy: Take the last time of the most recent chunk
		// in the series ***REMOVED***le. Then ***REMOVED***nd the oldest chunk among those
		// from heads.db that has a ***REMOVED***rst time later or equal to the
		// last time from the series ***REMOVED***le. Throw away the older chunks
		// from heads.db and stitch the parts together.

		// First, throw away the chunkDescs without chunks.
		s.chunkDescs = s.chunkDescs[s.persistWatermark:]
		chunk.NumMemDescs.Sub(float64(s.persistWatermark))
		cds, err := p.loadChunkDescs(fp, 0)
		if err != nil {
			log.Errorf(
				"Failed to load chunk descriptors for metric %v, ***REMOVED***ngerprint %v: %s",
				s.metric, fp, err,
			)
			purge()
			return fp, false
		}
		s.persistWatermark = len(cds)
		s.chunkDescsOffset = 0
		s.savedFirstTime = cds[0].FirstTime()
		s.modTime = modTime

		lastTime, err := cds[len(cds)-1].LastTime()
		if err != nil {
			log.Errorf(
				"Failed to determine time of the last sample for metric %v, ***REMOVED***ngerprint %v: %s",
				s.metric, fp, err,
			)
			purge()
			return fp, false
		}
		keepIdx := -1
		for i, cd := range s.chunkDescs {
			if cd.FirstTime() >= lastTime {
				keepIdx = i
				break
			}
		}
		if keepIdx == -1 {
			log.Warnf(
				"Recovered metric %v, ***REMOVED***ngerprint %v: all %d chunks recovered from series ***REMOVED***le.",
				s.metric, fp, chunksInFile,
			)
			chunk.NumMemDescs.Sub(float64(len(s.chunkDescs)))
			atomic.AddInt64(&chunk.NumMemChunks, int64(-len(s.chunkDescs)))
			s.chunkDescs = cds
			s.headChunkClosed = true
			// Finally, evict again all chunk.Descs except the latest one to save memory.
			s.evictChunkDescs(len(cds) - 1)
			return fp, true
		}
		log.Warnf(
			"Recovered metric %v, ***REMOVED***ngerprint %v: recovered %d chunks from series ***REMOVED***le, recovered %d chunks from checkpoint.",
			s.metric, fp, chunksInFile, len(s.chunkDescs)-keepIdx,
		)
		chunk.NumMemDescs.Sub(float64(keepIdx))
		atomic.AddInt64(&chunk.NumMemChunks, int64(-keepIdx))
		chunkDescsToEvict := len(cds)
		if keepIdx == len(s.chunkDescs) {
			// No chunks from series ***REMOVED***le left, head chunk is evicted, so declare it closed.
			s.headChunkClosed = true
			chunkDescsToEvict-- // Keep one chunk.Desc in this case to avoid a series with zero chunk.Descs.
		}
		s.chunkDescs = append(cds, s.chunkDescs[keepIdx:]...)
		// Finally, evict again chunk.Descs without chunk to save memory.
		s.evictChunkDescs(chunkDescsToEvict)
		return fp, true
	}
	// This series is supposed to be archived.
	metric, err := p.archivedMetric(fp)
	if err != nil {
		log.Errorf(
			"Fingerprint %v assumed archived but couldn't be looked up in archived index: %s",
			fp, err,
		)
		purge()
		return fp, false
	}
	if metric == nil {
		log.Warnf(
			"Fingerprint %v assumed archived but couldn't be found in archived index.",
			fp,
		)
		purge()
		return fp, false
	}
	// This series looks like a properly archived one.
	maybeAddMapping(fp, metric, fpm)
	return fp, true
}

func (p *persistence) cleanUpArchiveIndexes(
	fpToSeries map[model.Fingerprint]*memorySeries,
	fpsSeen map[model.Fingerprint]struct{},
	fpm fpMappings,
) error {
	log.Info("Cleaning up archive indexes.")
	var fp codable.Fingerprint
	var m codable.Metric
	count := 0
	if err := p.archivedFingerprintToMetrics.ForEach(func(kv index.KeyValueAccessor) error {
		count++
		if count%10000 == 0 {
			log.Infof("%d archived metrics checked.", count)
		}
		if err := kv.Key(&fp); err != nil {
			return err
		}
		_, fpSeen := fpsSeen[model.Fingerprint(fp)]
		inMemory := false
		if fpSeen {
			_, inMemory = fpToSeries[model.Fingerprint(fp)]
		}
		if !fpSeen || inMemory {
			if inMemory {
				log.Warnf("Archive clean-up: Fingerprint %v is not archived. Purging from archive indexes.", model.Fingerprint(fp))
			}
			if !fpSeen {
				log.Warnf("Archive clean-up: Fingerprint %v is unknown. Purging from archive indexes.", model.Fingerprint(fp))
			}
			// It's ***REMOVED***ne if the fp is not in the archive indexes.
			if _, err := p.archivedFingerprintToMetrics.Delete(fp); err != nil {
				return err
			}
			// Delete from timerange index, too.
			_, err := p.archivedFingerprintToTimeRange.Delete(fp)
			return err
		}
		// fp is legitimately archived. Now we need the metric to check for a mapped ***REMOVED***ngerprint.
		if err := kv.Value(&m); err != nil {
			return err
		}
		maybeAddMapping(model.Fingerprint(fp), model.Metric(m), fpm)
		// Make sure it is in timerange index, too.
		has, err := p.archivedFingerprintToTimeRange.Has(fp)
		if err != nil {
			return err
		}
		if has {
			return nil // All good.
		}
		log.Warnf("Archive clean-up: Fingerprint %v is not in time-range index. Unarchiving it for recovery.")
		// Again, it's ***REMOVED***ne if fp is not in the archive index.
		if _, err := p.archivedFingerprintToMetrics.Delete(fp); err != nil {
			return err
		}
		cds, err := p.loadChunkDescs(model.Fingerprint(fp), 0)
		if err != nil {
			return err
		}
		series, err := newMemorySeries(model.Metric(m), cds, p.seriesFileModTime(model.Fingerprint(fp)))
		if err != nil {
			return err
		}
		fpToSeries[model.Fingerprint(fp)] = series
		// Evict all but one chunk.Desc to save memory.
		series.evictChunkDescs(len(cds) - 1)
		return nil
	}); err != nil {
		return err
	}
	count = 0
	if err := p.archivedFingerprintToTimeRange.ForEach(func(kv index.KeyValueAccessor) error {
		count++
		if count%10000 == 0 {
			log.Infof("%d archived time ranges checked.", count)
		}
		if err := kv.Key(&fp); err != nil {
			return err
		}
		has, err := p.archivedFingerprintToMetrics.Has(fp)
		if err != nil {
			return err
		}
		if has {
			return nil // All good.
		}
		log.Warnf("Archive clean-up: Purging unknown ***REMOVED***ngerprint %v in time-range index.", fp)
		deleted, err := p.archivedFingerprintToTimeRange.Delete(fp)
		if err != nil {
			return err
		}
		if !deleted {
			log.Errorf("Fingerprint %v to be deleted from archivedFingerprintToTimeRange not found. This should never happen.", fp)
		}
		return nil
	}); err != nil {
		return err
	}
	log.Info("Clean-up of archive indexes complete.")
	return nil
}

func (p *persistence) rebuildLabelIndexes(
	fpToSeries map[model.Fingerprint]*memorySeries,
) error {
	count := 0
	log.Info("Rebuilding label indexes.")
	log.Info("Indexing metrics in memory.")
	for fp, s := range fpToSeries {
		p.indexMetric(fp, s.metric)
		count++
		if count%10000 == 0 {
			log.Infof("%d metrics queued for indexing.", count)
		}
	}
	log.Info("Indexing archived metrics.")
	var fp codable.Fingerprint
	var m codable.Metric
	if err := p.archivedFingerprintToMetrics.ForEach(func(kv index.KeyValueAccessor) error {
		if err := kv.Key(&fp); err != nil {
			return err
		}
		if err := kv.Value(&m); err != nil {
			return err
		}
		p.indexMetric(model.Fingerprint(fp), model.Metric(m))
		count++
		if count%10000 == 0 {
			log.Infof("%d metrics queued for indexing.", count)
		}
		return nil
	}); err != nil {
		return err
	}
	log.Info("All requests for rebuilding the label indexes queued. (Actual processing may lag behind.)")
	return nil
}

// maybeAddMapping adds a ***REMOVED***ngerprint mapping to fpm if the FastFingerprint of m is different from fp.
func maybeAddMapping(fp model.Fingerprint, m model.Metric, fpm fpMappings) {
	if rawFP := m.FastFingerprint(); rawFP != fp {
		log.Warnf(
			"Metric %v with ***REMOVED***ngerprint %v is mapped from raw ***REMOVED***ngerprint %v.",
			m, fp, rawFP,
		)
		if mappedFPs, ok := fpm[rawFP]; ok {
			mappedFPs[metricToUniqueString(m)] = fp
		} ***REMOVED*** {
			fpm[rawFP] = map[string]model.Fingerprint{
				metricToUniqueString(m): fp,
			}
		}
	}
}
