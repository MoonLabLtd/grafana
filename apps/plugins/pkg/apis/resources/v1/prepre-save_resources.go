package resourcespreserve

import (
	"fmt"
	"strings"
)

type ResourceKind int

const (
	ResourceKindDatasources ResourceKind = 0
	ResourceKindDatabases   ResourceKind = 1
	ResourceKindWorkgroups  ResourceKind = 2
)

func resourceKindName(kind ResourceKind) string {
	switch kind {
	case ResourceKindDatasources:
		return "datasources"
	case ResourceKindDatabases:
		return "databases"
	case ResourceKindWorkgroups:
		return "workgroups"
	default:
		return "unknown"
	}
}

type ResourceRef struct {
	Label string
	Value string
}

type ResourceContext struct {
	Region        string
	AuthType      string
	AssumeRoleArn string
	ExternalId    string
	DataSource    string
	Database      string
	Workgroup     string
}

type ResourceResult struct {
	Resources []ResourceRef
	Empty     bool
}

type ResourceError struct {
	Code    string
	Message string
}

type ResourceFetcher interface {
	Fetch(ctx ResourceContext, kind ResourceKind) ([]ResourceRef, error)
}

type PresaveResourceService struct {
	fetcher ResourceFetcher
}

func NewPresaveResourceService(fetcher ResourceFetcher) *PresaveResourceService {
	return &PreSaveResourceService{fetcher: fetcher}
}

func validateContext(ctx ResourceContext) *ResourceError {
	if strings.TrimSpace(ctx.Region) == "" {
		return &ResourceError{Code: "InvalidContext", Message: "region is required"}
	}
	if strings.TrimSpace(ctx.AuthType) == "" {
		return &ResourceError{Code: "InvalidContext", Message: "authType is required"}
	}
	return nil
}

func (s *PresaveResourceService) Resolve(ctx ResourceContext, kind ResourceKind) (ResourceResult, *ResourceError) {
	if err := validateContext(ctx); err != nil {
		return ResourceResult{}, err
	}
	refs, fetchErr := s.fetcher.Fetch(ctx, kind)
	if fetchErr != nil {
		return ResourceResult{}, &ResourceError{
			Code:    "AwsLookupFailed",
			Message: fmt.Sprintf("%s: %s", resourceKindName(kind), fetchErr.Error()),
		}
	}
	return ResourceResult{Resources: refs, Empty: len(refs) == 0}, nil
}

type RequiredFieldError struct {
	Field   string
	Message string
}

func (s *PresaveResourceService) ValidateRequiredFields(ctx ResourceContext) []RequiredFieldError {
	var result []RequiredFieldError
	if strings.TrimSpace(ctx.DataSource) == "" {
		result = append(result, RequiredFieldError{Field: "dataSource", Message: "Data source is required"})
	}
	if strings.TrimSpace(ctx.Database) == "" {
		result = append(result, RequiredFieldError{Field: "database", Message: "Database is required"})
	}
	if strings.TrimSpace(ctx.Workgroup) == "" {
		result = append(result, RequiredFieldError{Field: "workgroup", Message: "Workgroup is required"})
	}
	return result
}

func (s *PresaveResourceService) RejectEmptyWorkgroup(workgroup string) *ResourceError {
	if strings.TrimSpace(workgroup) == "" {
		return &ResourceError{Code: "InvalidWorkgroup", Message: "Workgroup is required"}
	}
	return nil
}

func (s *PresaveResourceService) LookupFailureWarning(kind ResourceKind, instanceUid string, lookupErr error) string {
	code := "unknown"
	if lookupErr != nil {
		code = lookupErr.Error()
	}
	return fmt.Sprintf("pre-save resource lookup failed kind=%s instance=%s code=%s", resourceKindName(kind), instanceUid, code)
}
