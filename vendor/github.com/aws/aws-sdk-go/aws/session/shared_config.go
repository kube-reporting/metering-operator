package session

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/internal/ini"
)

const (
	// Static Credentials group
	accessKeyIDKey  = `aws_access_key_id`     // group required
	secretAccessKey = `aws_secret_access_key` // group required
	sessionTokenKey = `aws_session_token`     // optional

	// Assume Role Credentials group
	roleArnKey          = `role_arn`          // group required
	sourcePro***REMOVED***leKey    = `source_pro***REMOVED***le`    // group required (or credential_source)
	credentialSourceKey = `credential_source` // group required (or source_pro***REMOVED***le)
	externalIDKey       = `external_id`       // optional
	mfaSerialKey        = `mfa_serial`        // optional
	roleSessionNameKey  = `role_session_name` // optional

	// CSM options
	csmEnabledKey  = `csm_enabled`
	csmHostKey     = `csm_host`
	csmPortKey     = `csm_port`
	csmClientIDKey = `csm_client_id`

	// Additional Con***REMOVED***g ***REMOVED***elds
	regionKey = `region`

	// endpoint discovery group
	enableEndpointDiscoveryKey = `endpoint_discovery_enabled` // optional

	// External Credential Process
	credentialProcessKey = `credential_process` // optional

	// Web Identity Token File
	webIdentityTokenFileKey = `web_identity_token_***REMOVED***le` // optional

	// Additional con***REMOVED***g ***REMOVED***elds for regional or legacy endpoints
	stsRegionalEndpointSharedKey = `sts_regional_endpoints`

	// DefaultSharedCon***REMOVED***gPro***REMOVED***le is the default pro***REMOVED***le to be used when
	// loading con***REMOVED***guration from the con***REMOVED***g ***REMOVED***les if another pro***REMOVED***le name
	// is not provided.
	DefaultSharedCon***REMOVED***gPro***REMOVED***le = `default`
)

// sharedCon***REMOVED***g represents the con***REMOVED***guration ***REMOVED***elds of the SDK con***REMOVED***g ***REMOVED***les.
type sharedCon***REMOVED***g struct {
	// Credentials values from the con***REMOVED***g ***REMOVED***le. Both aws_access_key_id and
	// aws_secret_access_key must be provided together in the same ***REMOVED***le to be
	// considered valid. The values will be ignored if not a complete group.
	// aws_session_token is an optional ***REMOVED***eld that can be provided if both of
	// the other two ***REMOVED***elds are also provided.
	//
	//	aws_access_key_id
	//	aws_secret_access_key
	//	aws_session_token
	Creds credentials.Value

	CredentialSource     string
	CredentialProcess    string
	WebIdentityTokenFile string

	RoleARN         string
	RoleSessionName string
	ExternalID      string
	MFASerial       string

	SourcePro***REMOVED***leName string
	SourcePro***REMOVED***le     *sharedCon***REMOVED***g

	// Region is the region the SDK should use for looking up AWS service
	// endpoints and signing requests.
	//
	//	region
	Region string

	// EnableEndpointDiscovery can be enabled in the shared con***REMOVED***g by setting
	// endpoint_discovery_enabled to true
	//
	//	endpoint_discovery_enabled = true
	EnableEndpointDiscovery *bool
	// CSM Options
	CSMEnabled  *bool
	CSMHost     string
	CSMPort     string
	CSMClientID string

	// Speci***REMOVED***es the Regional Endpoint flag for the sdk to resolve the endpoint for a service
	//
	// sts_regional_endpoints = sts_regional_endpoint
	// This can take value as `LegacySTSEndpoint` or `RegionalSTSEndpoint`
	STSRegionalEndpoint endpoints.STSRegionalEndpoint
}

type sharedCon***REMOVED***gFile struct {
	Filename string
	IniData  ini.Sections
}

// loadSharedCon***REMOVED***g retrieves the con***REMOVED***guration from the list of ***REMOVED***les using
// the pro***REMOVED***le provided. The order the ***REMOVED***les are listed will determine
// precedence. Values in subsequent ***REMOVED***les will overwrite values de***REMOVED***ned in
// earlier ***REMOVED***les.
//
// For example, given two ***REMOVED***les A and B. Both de***REMOVED***ne credentials. If the order
// of the ***REMOVED***les are A then B, B's credential values will be used instead of
// A's.
//
// See sharedCon***REMOVED***g.setFromFile for information how the con***REMOVED***g ***REMOVED***les
// will be loaded.
func loadSharedCon***REMOVED***g(pro***REMOVED***le string, ***REMOVED***lenames []string, exOpts bool) (sharedCon***REMOVED***g, error) {
	if len(pro***REMOVED***le) == 0 {
		pro***REMOVED***le = DefaultSharedCon***REMOVED***gPro***REMOVED***le
	}

	***REMOVED***les, err := loadSharedCon***REMOVED***gIniFiles(***REMOVED***lenames)
	if err != nil {
		return sharedCon***REMOVED***g{}, err
	}

	cfg := sharedCon***REMOVED***g{}
	pro***REMOVED***les := map[string]struct{}{}
	if err = cfg.setFromIniFiles(pro***REMOVED***les, pro***REMOVED***le, ***REMOVED***les, exOpts); err != nil {
		return sharedCon***REMOVED***g{}, err
	}

	return cfg, nil
}

func loadSharedCon***REMOVED***gIniFiles(***REMOVED***lenames []string) ([]sharedCon***REMOVED***gFile, error) {
	***REMOVED***les := make([]sharedCon***REMOVED***gFile, 0, len(***REMOVED***lenames))

	for _, ***REMOVED***lename := range ***REMOVED***lenames {
		sections, err := ini.OpenFile(***REMOVED***lename)
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == ini.ErrCodeUnableToReadFile {
			// Skip ***REMOVED***les which can't be opened and read for whatever reason
			continue
		} ***REMOVED*** if err != nil {
			return nil, SharedCon***REMOVED***gLoadError{Filename: ***REMOVED***lename, Err: err}
		}

		***REMOVED***les = append(***REMOVED***les, sharedCon***REMOVED***gFile{
			Filename: ***REMOVED***lename, IniData: sections,
		})
	}

	return ***REMOVED***les, nil
}

func (cfg *sharedCon***REMOVED***g) setFromIniFiles(pro***REMOVED***les map[string]struct{}, pro***REMOVED***le string, ***REMOVED***les []sharedCon***REMOVED***gFile, exOpts bool) error {
	// Trim ***REMOVED***les from the list that don't exist.
	var skippedFiles int
	var pro***REMOVED***leNotFoundErr error
	for _, f := range ***REMOVED***les {
		if err := cfg.setFromIniFile(pro***REMOVED***le, f, exOpts); err != nil {
			if _, ok := err.(SharedCon***REMOVED***gPro***REMOVED***leNotExistsError); ok {
				// Ignore pro***REMOVED***les not de***REMOVED***ned in individual ***REMOVED***les.
				pro***REMOVED***leNotFoundErr = err
				skippedFiles++
				continue
			}
			return err
		}
	}
	if skippedFiles == len(***REMOVED***les) {
		// If all ***REMOVED***les were skipped because the pro***REMOVED***le is not found, return
		// the original pro***REMOVED***le not found error.
		return pro***REMOVED***leNotFoundErr
	}

	if _, ok := pro***REMOVED***les[pro***REMOVED***le]; ok {
		// if this is the second instance of the pro***REMOVED***le the Assume Role
		// options must be cleared because they are only valid for the
		// ***REMOVED***rst reference of a pro***REMOVED***le. The self linked instance of the
		// pro***REMOVED***le only have credential provider options.
		cfg.clearAssumeRoleOptions()
	} ***REMOVED*** {
		// First time a pro***REMOVED***le has been seen, It must either be a assume role
		// or credentials. Assert if the credential type requires a role ARN,
		// the ARN is also set.
		if err := cfg.validateCredentialsRequireARN(pro***REMOVED***le); err != nil {
			return err
		}
	}
	pro***REMOVED***les[pro***REMOVED***le] = struct{}{}

	if err := cfg.validateCredentialType(); err != nil {
		return err
	}

	// Link source pro***REMOVED***les for assume roles
	if len(cfg.SourcePro***REMOVED***leName) != 0 {
		// Linked pro***REMOVED***le via source_pro***REMOVED***le ignore credential provider
		// options, the source pro***REMOVED***le must provide the credentials.
		cfg.clearCredentialOptions()

		srcCfg := &sharedCon***REMOVED***g{}
		err := srcCfg.setFromIniFiles(pro***REMOVED***les, cfg.SourcePro***REMOVED***leName, ***REMOVED***les, exOpts)
		if err != nil {
			// SourcePro***REMOVED***le that doesn't exist is an error in con***REMOVED***guration.
			if _, ok := err.(SharedCon***REMOVED***gPro***REMOVED***leNotExistsError); ok {
				err = SharedCon***REMOVED***gAssumeRoleError{
					RoleARN:       cfg.RoleARN,
					SourcePro***REMOVED***le: cfg.SourcePro***REMOVED***leName,
				}
			}
			return err
		}

		if !srcCfg.hasCredentials() {
			return SharedCon***REMOVED***gAssumeRoleError{
				RoleARN:       cfg.RoleARN,
				SourcePro***REMOVED***le: cfg.SourcePro***REMOVED***leName,
			}
		}

		cfg.SourcePro***REMOVED***le = srcCfg
	}

	return nil
}

// setFromFile loads the con***REMOVED***guration from the ***REMOVED***le using the pro***REMOVED***le
// provided. A sharedCon***REMOVED***g pointer type value is used so that multiple con***REMOVED***g
// ***REMOVED***le loadings can be chained.
//
// Only loads complete logically grouped values, and will not set ***REMOVED***elds in cfg
// for incomplete grouped values in the con***REMOVED***g. Such as credentials. For
// example if a con***REMOVED***g ***REMOVED***le only includes aws_access_key_id but no
// aws_secret_access_key the aws_access_key_id will be ignored.
func (cfg *sharedCon***REMOVED***g) setFromIniFile(pro***REMOVED***le string, ***REMOVED***le sharedCon***REMOVED***gFile, exOpts bool) error {
	section, ok := ***REMOVED***le.IniData.GetSection(pro***REMOVED***le)
	if !ok {
		// Fallback to to alternate pro***REMOVED***le name: pro***REMOVED***le <name>
		section, ok = ***REMOVED***le.IniData.GetSection(fmt.Sprintf("pro***REMOVED***le %s", pro***REMOVED***le))
		if !ok {
			return SharedCon***REMOVED***gPro***REMOVED***leNotExistsError{Pro***REMOVED***le: pro***REMOVED***le, Err: nil}
		}
	}

	if exOpts {
		// Assume Role Parameters
		updateString(&cfg.RoleARN, section, roleArnKey)
		updateString(&cfg.ExternalID, section, externalIDKey)
		updateString(&cfg.MFASerial, section, mfaSerialKey)
		updateString(&cfg.RoleSessionName, section, roleSessionNameKey)
		updateString(&cfg.SourcePro***REMOVED***leName, section, sourcePro***REMOVED***leKey)
		updateString(&cfg.CredentialSource, section, credentialSourceKey)
		updateString(&cfg.Region, section, regionKey)

		if v := section.String(stsRegionalEndpointSharedKey); len(v) != 0 {
			sre, err := endpoints.GetSTSRegionalEndpoint(v)
			if err != nil {
				return fmt.Errorf("failed to load %s from shared con***REMOVED***g, %s, %v",
					stsRegionalEndpointKey, ***REMOVED***le.Filename, err)
			}
			cfg.STSRegionalEndpoint = sre
		}
	}

	updateString(&cfg.CredentialProcess, section, credentialProcessKey)
	updateString(&cfg.WebIdentityTokenFile, section, webIdentityTokenFileKey)

	// Shared Credentials
	creds := credentials.Value{
		AccessKeyID:     section.String(accessKeyIDKey),
		SecretAccessKey: section.String(secretAccessKey),
		SessionToken:    section.String(sessionTokenKey),
		ProviderName:    fmt.Sprintf("SharedCon***REMOVED***gCredentials: %s", ***REMOVED***le.Filename),
	}
	if creds.HasKeys() {
		cfg.Creds = creds
	}

	// Endpoint discovery
	updateBoolPtr(&cfg.EnableEndpointDiscovery, section, enableEndpointDiscoveryKey)

	// CSM options
	updateBoolPtr(&cfg.CSMEnabled, section, csmEnabledKey)
	updateString(&cfg.CSMHost, section, csmHostKey)
	updateString(&cfg.CSMPort, section, csmPortKey)
	updateString(&cfg.CSMClientID, section, csmClientIDKey)

	return nil
}

func (cfg *sharedCon***REMOVED***g) validateCredentialsRequireARN(pro***REMOVED***le string) error {
	var credSource string

	switch {
	case len(cfg.SourcePro***REMOVED***leName) != 0:
		credSource = sourcePro***REMOVED***leKey
	case len(cfg.CredentialSource) != 0:
		credSource = credentialSourceKey
	case len(cfg.WebIdentityTokenFile) != 0:
		credSource = webIdentityTokenFileKey
	}

	if len(credSource) != 0 && len(cfg.RoleARN) == 0 {
		return CredentialRequiresARNError{
			Type:    credSource,
			Pro***REMOVED***le: pro***REMOVED***le,
		}
	}

	return nil
}

func (cfg *sharedCon***REMOVED***g) validateCredentialType() error {
	// Only one or no credential type can be de***REMOVED***ned.
	if !oneOrNone(
		len(cfg.SourcePro***REMOVED***leName) != 0,
		len(cfg.CredentialSource) != 0,
		len(cfg.CredentialProcess) != 0,
		len(cfg.WebIdentityTokenFile) != 0,
	) {
		return ErrSharedCon***REMOVED***gSourceCollision
	}

	return nil
}

func (cfg *sharedCon***REMOVED***g) hasCredentials() bool {
	switch {
	case len(cfg.SourcePro***REMOVED***leName) != 0:
	case len(cfg.CredentialSource) != 0:
	case len(cfg.CredentialProcess) != 0:
	case len(cfg.WebIdentityTokenFile) != 0:
	case cfg.Creds.HasKeys():
	default:
		return false
	}

	return true
}

func (cfg *sharedCon***REMOVED***g) clearCredentialOptions() {
	cfg.CredentialSource = ""
	cfg.CredentialProcess = ""
	cfg.WebIdentityTokenFile = ""
	cfg.Creds = credentials.Value{}
}

func (cfg *sharedCon***REMOVED***g) clearAssumeRoleOptions() {
	cfg.RoleARN = ""
	cfg.ExternalID = ""
	cfg.MFASerial = ""
	cfg.RoleSessionName = ""
	cfg.SourcePro***REMOVED***leName = ""
}

func oneOrNone(bs ...bool) bool {
	var count int

	for _, b := range bs {
		if b {
			count++
			if count > 1 {
				return false
			}
		}
	}

	return true
}

// updateString will only update the dst with the value in the section key, key
// is present in the section.
func updateString(dst *string, section ini.Section, key string) {
	if !section.Has(key) {
		return
	}
	*dst = section.String(key)
}

// updateBoolPtr will only update the dst with the value in the section key,
// key is present in the section.
func updateBoolPtr(dst **bool, section ini.Section, key string) {
	if !section.Has(key) {
		return
	}
	*dst = new(bool)
	**dst = section.Bool(key)
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
	RoleARN       string
	SourcePro***REMOVED***le string
}

// Code is the short id of the error.
func (e SharedCon***REMOVED***gAssumeRoleError) Code() string {
	return "SharedCon***REMOVED***gAssumeRoleError"
}

// Message is the description of the error
func (e SharedCon***REMOVED***gAssumeRoleError) Message() string {
	return fmt.Sprintf(
		"failed to load assume role for %s, source pro***REMOVED***le %s has no shared credentials",
		e.RoleARN, e.SourcePro***REMOVED***le,
	)
}

// OrigErr is the underlying error that caused the failure.
func (e SharedCon***REMOVED***gAssumeRoleError) OrigErr() error {
	return nil
}

// Error satis***REMOVED***es the error interface.
func (e SharedCon***REMOVED***gAssumeRoleError) Error() string {
	return awserr.SprintError(e.Code(), e.Message(), "", nil)
}

// CredentialRequiresARNError provides the error for shared con***REMOVED***g credentials
// that are incorrectly con***REMOVED***gured in the shared con***REMOVED***g or credentials ***REMOVED***le.
type CredentialRequiresARNError struct {
	// type of credentials that were con***REMOVED***gured.
	Type string

	// Pro***REMOVED***le name the credentials were in.
	Pro***REMOVED***le string
}

// Code is the short id of the error.
func (e CredentialRequiresARNError) Code() string {
	return "CredentialRequiresARNError"
}

// Message is the description of the error
func (e CredentialRequiresARNError) Message() string {
	return fmt.Sprintf(
		"credential type %s requires role_arn, pro***REMOVED***le %s",
		e.Type, e.Pro***REMOVED***le,
	)
}

// OrigErr is the underlying error that caused the failure.
func (e CredentialRequiresARNError) OrigErr() error {
	return nil
}

// Error satis***REMOVED***es the error interface.
func (e CredentialRequiresARNError) Error() string {
	return awserr.SprintError(e.Code(), e.Message(), "", nil)
}
