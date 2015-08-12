package main

import ()

type Metadata struct {
	Guid string `json:"guid"`
	Url  string `json:"url,omitempty"`
}
type ApplicationEntity struct {
	Name               string            `json:"name,omitempty"`
	Command            string            `json:"command,omitempty"`
	State              string            `json:"state,omitempty"`
	SpaceGuid          string            `json:"space_guid,omitempty"`
	Instances          int               `json:"instances,omitempty"`
	Memory             uint64            `json:"memory,omitempty"`
	DiskQuota          uint64            `json:"disk_quota,omitempty"`
	StackGuid          string            `json:"stack_guid,omitempty"`
	Buildpack          string            `json:"buildpack,omitempty"`
	EnvironmentJson    map[string]string `json:"environment_json,omitempty"`
	HealthCheckTimeout int               `json:"health_check_timeout,omitempty"`
}
type ApplicationResource struct {
	Metadata Metadata
	Entity   ApplicationEntity
}
type PaginatedApplicationResources struct {
	Resources []ApplicationResource
}
