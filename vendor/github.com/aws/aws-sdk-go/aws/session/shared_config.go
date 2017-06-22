package session

import (
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/go-ini/ini"
)

const (
	// Static Credentials group
	accessKeyIDKey  = `aws_access_key_id`     // group required
	secretAccessKey = `aws_secret_access_key` // group required
	sessionTokenKey = `aws_session_token`     // optional

	// Assume Role Credentials group
	roleArnKey         = `role_arn`          // group required
	sourcePro***REMOVED***leKey   = `source_pro***REMOVED***le`    // group required
	externalIDKey      = `external_id`       // optional
	mfaSerialKey       = `mfa_serial`        // optional
	roleSessionNameKey = `role_session_name` // optional

	// Additional Con***REMOVED***g ***REMOVED***elds
	regionKey = `region`

	// DefaultSharedCon***REMOVED***gPro***REMOVED***le is the default pro***REMOVED***le to be used when
	// loading con***REMOVED***guration from the con***REMOVED***g ***REMOVED***les if another pro***REMOVED***le name
	// is not provided.
	DefaultSharedCon***REMOVED***gPro***REMOVED***le = `default`
)

type assumeRoleCon***REMOVED***g struct {
	RoleARN         string
	SourcePro***REMOVED***le   string
	ExternalID      string
	MFASerial       string
	RoleSessionName string
}

// sharedCon***REMOVED***g represents the con***REMOVED***guration ***REMOVED***elds of the SDK con***REMOVED***g ***REMOVED***les.
type sharedCon***REMOVED***g struct {
	// Credentials values from the con***REMOVED***g ***REMOVED***le. Both aws_access_key_id
	// and aws_secret_access_key must be provided together in the same ***REMOVED***le
	// to be considered valid. The values will be ignored if not a complete group.
	// aws_session_token is an optional ***REMOVED***eld that can be provided if both of the
	// other two ***REMOVED***elds are also provided.
	//
	//	aws_access_key_id
	//	aws_secret_access_key
	//	aws_session_token
	Creds credentials.Value

	AssumeRole       assumeRoleCon***REMOVED***g
	AssumeRoleSource *sharedCon***REMOVED***g

	// Region is the region the SDK should use for looking up AWS service endpoints
	// and signing requests.
	//
	//	region
	Region string
}

type sharedCon***REMOVED***gFile struct {
	Filename string
	IniData  *ini.File
}

// loadSharedCon***REMOVED***g retrieves the con***REMOVED***guration from the list of ***REMOVED***les
// using the pro***REMOVED***le provided. The order the ***REMOVED***les are listed will determine
// precedence. Values in subsequent ***REMOVED***les will overwrite values de***REMOVED***ned in
// earlier ***REMOVED***les.
//
// For example, given two ***REMOVED***les A and B. Both de***REMOVED***ne credentials. If the order
// of the ***REMOVED***les are A then B, B's credential values will be used instead of A's.
//
// See sharedCon***REMOVED***g.setFromFile for information how the con***REMOVED***g ***REMOVED***les
// will be loaded.
func loadSharedCon***REMOVED***g(pro***REMOVED***le string, ***REMOVED***lenames []string) (sharedCon***REMOVED***g, error) {
	if len(pro***REMOVED***le) == 0 {
		pro***REMOVED***le = DefaultSharedCon***REMOVED***gPro***REMOVED***le
	}

	***REMOVED***les, err := loadSharedCon***REMOVED***gIniFiles(***REMOVED***lenames)
	if err != nil {
		return sharedCon***REMOVED***g{}, err
	}

	cfg := sharedCon***REMOVED***g{}
	if err = cfg.setFromIniFiles(pro***REMOVED***le, ***REMOVED***les); err != nil {
		return sharedCon***REMOVED***g{}, err
	}

	if len(cfg.AssumeRole.SourcePro***REMOVED***le) > 0 {
		if err := cfg.setAssumeRoleSource(pro***REMOVED***le, ***REMOVED***les); err != nil {
			return sharedCon***REMOVED***g{}, err
		}
	}

	return cfg, nil
}

func loadSharedCon***REMOVED***gIniFiles(***REMOVED***lenames []string) ([]sharedCon***REMOVED***gFile, error) {
	***REMOVED***les := make([]sharedCon***REMOVED***gFile, 0, len(***REMOVED***lenames))

	for _, ***REMOVED***lename := range ***REMOVED***lenames {
		b, err := ioutil.ReadFile(***REMOVED***lename)
		if err != nil {
			// Skip ***REMOVED***les which can't be opened and read for whatever reason
			continue
		}

		f, err := ini.Load(b)
		if err != nil {
			return nil, SharedCon***REMOVED***gLoadError{Filename: ***REMOVED***lename, Err: err}
		}

		***REMOVED***les = append(***REMOVED***les, sharedCon***REMOVED***gFile{
			Filename: ***REMOVED***lename, IniData: f,
		})
	}

	return ***REMOVED***les, nil
}

func (cfg *sharedCon***REMOVED***g) setAssumeRoleSource(origPro***REMOVED***le string, ***REMOVED***les []sharedCon***REMOVED***gFile) error {
	var assumeRoleSrc sharedCon***REMOVED***g

	// Multiple level assume role chains are not support
	if cfg.AssumeRole.SourcePro***REMOVED***le == origPro***REMOVED***le {
		assumeRoleSrc = *cfg
		assumeRoleSrc.AssumeRole = assumeRoleCon***REMOVED***g{}
	} ***REMOVED*** {
		err := assumeRoleSrc.setFromIniFiles(cfg.AssumeRole.SourcePro***REMOVED***le, ***REMOVED***les)
		if err != nil {
			return err
		}
	}

	if len(assumeRoleSrc.Creds.AccessKeyID) == 0 {
		return SharedCon***REMOVED***gAssumeRoleError{RoleARN: cfg.AssumeRole.RoleARN}
	}

	cfg.AssumeRoleSource = &assumeRoleSrc

	return nil
}

func (cfg *sharedCon***REMOVED***g) setFromIniFiles(pro***REMOVED***le string, ***REMOVED***les []sharedCon***REMOVED***gFile) error {
	// Trim ***REMOVED***les from the list that don't exist.
	for _, f := range ***REMOVED***les {
		if err := cfg.setFromIniFile(pro***REMOVED***le, f); err != nil {
			if _, ok := err.(SharedCon***REMOVED***gPro***REMOVED***leNotExistsError); ok {
				// Ignore proviles missings
				continue
			}
			return err
		}
	}

	return nil
}

// setFromFile loads the con***REMOVED***guration from the ***REMOVED***le using
// the pro***REMOVED***le provided. A sharedCon***REMOVED***g pointer type value is used so that
// multiple con***REMOVED***g ***REMOVED***le loadings can be chained.
//
// Only loads complete logically grouped values, and will not set ***REMOVED***elds in cfg
// for incomplete grouped values in the con***REMOVED***g. Such as credentials. For example
// if a con***REMOVED***g ***REMOVED***le only includes aws_access_key_id but no aws_secret_access_key
// the aws_access_key_id will be ignored.
func (cfg *sharedCon***REMOVED***g) setFromIniFile(pro***REMOVED***le string, ***REMOVED***le sharedCon***REMOVED***gFile) error {
	section, err := ***REMOVED***le.IniData.GetSection(pro***REMOVED***le)
	if err != nil {
		// Fallback to to alternate pro***REMOVED***le name: pro***REMOVED***le <name>
		section, err = ***REMOVED***le.IniData.GetSection(fmt.Sprintf("pro***REMOVED***le %s", pro***REMOVED***le))
		if err != nil {
			return SharedCon***REMOVED***gPro***REMOVED***leNotExistsError{Pro***REMOVED***le: pro***REMOVED***le, Err: err}
		}
	}

	// Shared Credentials
	akid := section.Key(accessKeyIDKey).String()
	secret := section.Key(secretAccessKey).String()
	if len(akid) > 0 && len(secret) > 0 {
		cfg.Creds = credentials.Value{
			AccessKeyID:     akid,
			SecretAccessKey: secret,
			SessionToken:    section.Key(sessionTokenKey).String(),
			ProviderName:    fmt.Sprintf("SharedCon***REMOVED***gCredentials: %s", ***REMOVED***le.Filename),
		}
	}

	// Assume Role
	roleArn := section.Key(roleArnKey).String()
	srcPro***REMOVED***le := section.Key(sourcePro***REMOVED***leKey).String()
	if len(roleArn) > 0 && len(srcPro***REMOVED***le) > 0 {
		cfg.AssumeRole = assumeRoleCon***REMOVED***g{
			RoleARN:         roleArn,
			SourcePro***REMOVED***le:   srcPro***REMOVED***le,
			ExternalID:      section.Key(externalIDKey).String(),
			MFASerial:       section.Key(mfaSerialKey).String(),
			RoleSessionName: section.Key(roleSessionNameKey).String(),
		}
	}

	// Region
	if v := section.Key(regionKey).String(); len(v) > 0 {
		cfg.Region = v
	}

	return nil
}

// SharedCon***REMOVED***gLoadError is an error for the shared con***REMOVED***g ***REMOVED***le failed to load.
type SharedCon***REMOVED***gLoadError struct {
	Filename string
	Err      error
}

// Code is the short id of the error.
func (e SharedCon***REMOVED***gLoadError) Code() string {
	return "SharedCon***REMOVED***gLoadError"
}

// Message is the description of the error
func (e SharedCon***REMOVED***gLoadError) Message() string {
	return fmt.Sprintf("failed to load con***REMOVED***g ***REMOVED***le, %s", e.Filename)
}

// OrigErr is the underlying error that caused the failure.
func (e SharedCon***REMOVED***gLoadError) OrigErr() error {
	return e.Err
}

// Error satis***REMOVED***es the error interface.
func (e SharedCon***REMOVED***gLoadError) Error() string {
	return awserr.SprintError(e.Code(), e.Message(), "", e.Err)
}

// SharedCon***REMOVED***gPro***REMOVED***leNotExistsError is an error for the shared con***REMOVED***g when
// the pro***REMOVED***le was not ***REMOVED***nd in the con***REMOVED***g ***REMOVED***le.
type SharedCon***REMOVED***gPro***REMOVED***leNotExistsError struct {
	Pro***REMOVED***le string
	Err     error
}

// Code is the short id of the error.
func (e SharedCon***REMOVED***gPro***REMOVED***leNotExistsError) Code() string {
	return "SharedCon***REMOVED***gPro***REMOVED***leNotExistsError"
}

// Message is the description of the error
func (e SharedCon***REMOVED***gPro***REMOVED***leNotExistsError) Message() string {
	return fmt.Sprintf("failed to get pro***REMOVED***le, %s", e.Pro***REMOVED***le)
}

// OrigErr is the underlying error that caused the failure.
func (e SharedCon***REMOVED***gPro***REMOVED***leNotExistsError) OrigErr() error {
	return e.Err
}

// Error satis***REMOVED***es the error interface.
func (e SharedCon***REMOVED***gPro***REMOVED***leNotExistsError) Error() string {
	return awserr.SprintError(e.Code(), e.Message(), "", e.Err)
}

// SharedCon***REMOVED***gAssumeRoleError is an error for the shared con***REMOVED***g when the
// pro***REMOVED***le contains assume role information, but that information is invalid
// or not complete.
type SharedCon***REMOVED***gAssumeRoleError struct {
	RoleARN string
}

// Code is the short id of the error.
func (e SharedCon***REMOVED***gAssumeRoleError) Code() string {
	return "SharedCon***REMOVED***gAssumeRoleError"
}

// Message is the description of the error
func (e SharedCon***REMOVED***gAssumeRoleError) Message() string {
	return fmt.Sprintf("failed to load assume role for %s, source pro***REMOVED***le has no shared credentials",
		e.RoleARN)
}

// OrigErr is the underlying error that caused the failure.
func (e SharedCon***REMOVED***gAssumeRoleError) OrigErr() error {
	return nil
}

// Error satis***REMOVED***es the error interface.
func (e SharedCon***REMOVED***gAssumeRoleError) Error() string {
	return awserr.SprintError(e.Code(), e.Message(), "", nil)
}
