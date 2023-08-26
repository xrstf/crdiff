// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package compare

import (
	"github.com/tufin/oasdiff/checker"
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/utils"

	"go.xrstf.de/crdiff/pkg/compare/oasdiff"
)

// Change is a generic change that is not schema-specific, e.g. when
// a CRD scope was changed.
type Change struct {
	Breaking    bool   `json:"breaking" yaml:"breaking"`
	Description string `json:"description,omitempty" yaml:"description"`
}

// CRDDiff describes all differences for all versions of a single CRD.
type CRDDiff struct {
	General         []Change                  `json:"generalChanges,omitempty" yaml:"generalChanges,omitempty"`
	AddedVersions   utils.StringList          `json:"added,omitempty" yaml:"added,omitempty"`
	DeletedVersions utils.StringList          `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	ChangedVersions map[string]CRDVersionDiff `json:"changed,omitempty" yaml:"changed,omitempty"`
}

func (d *CRDDiff) HasChanges() bool {
	if d == nil {
		return false
	}

	if len(d.General) > 0 {
		return true
	}

	if len(d.AddedVersions) > 0 {
		return true
	}

	if len(d.DeletedVersions) > 0 {
		return true
	}

	for _, v := range d.ChangedVersions {
		if v.HasChanges() {
			return true
		}
	}

	return false
}

func (d *CRDDiff) HasBreakingChanges() bool {
	if d == nil {
		return false
	}

	for _, c := range d.General {
		if c.Breaking {
			return true
		}
	}

	if d.DeletedVersions.Len() > 0 {
		return true
	}

	for _, versionDiff := range d.ChangedVersions {
		if versionDiff.HasBreakingChanges() {
			return true
		}
	}

	return false
}

func (in *CRDDiff) DeepCopy() *CRDDiff {
	if in == nil {
		return nil
	}

	out := &CRDDiff{
		General:         make([]Change, len(in.General)),
		AddedVersions:   make(utils.StringList, len(in.AddedVersions)),
		DeletedVersions: make(utils.StringList, len(in.DeletedVersions)),
		ChangedVersions: make(map[string]CRDVersionDiff, len(in.ChangedVersions)),
	}

	copy(out.General, in.General)
	copy(out.AddedVersions, in.AddedVersions)
	copy(out.DeletedVersions, in.DeletedVersions)

	for k, v := range in.ChangedVersions {
		out.ChangedVersions[k] = *v.DeepCopy()
	}

	return out
}

// CRDVersionDiff describes all schema changes for a single
// CRD version (e.g. all changes for core/v1's Pods).
// +k8s:deepcopy-gen=true
type CRDVersionDiff struct {
	SchemaChanges   map[string]CRDSchemaDiff `json:"schemaChanges,omitempty" yaml:"schemaChanges,omitempty"`
	BreakingChanges []BreakingChange         `json:"breakingChanges,omitempty" yaml:"breakingChanges,omitempty"`
}

func (d *CRDVersionDiff) HasChanges() bool {
	if d == nil {
		return false
	}

	return len(d.SchemaChanges) > 0
}

func (d *CRDVersionDiff) HasBreakingChanges() bool {
	if d == nil {
		return false
	}

	return len(d.BreakingChanges) > 0
}

func (in *CRDVersionDiff) DeepCopy() *CRDVersionDiff {
	if in == nil {
		return nil
	}

	out := &CRDVersionDiff{
		SchemaChanges:   make(map[string]CRDSchemaDiff, len(in.SchemaChanges)),
		BreakingChanges: make([]BreakingChange, len(in.BreakingChanges)),
	}

	copy(out.BreakingChanges, in.BreakingChanges)

	for k, v := range in.SchemaChanges {
		out.SchemaChanges[k] = *v.DeepCopy()
	}

	return out
}

// CRDSchemaDiff contains the changes at a given path within
// a single CRD's version schema (e.g. changes at spec.clusterRef in
// core/v1's Pod).
type CRDSchemaDiff struct {
	AddedProperties   utils.StringList `json:"added,omitempty" yaml:"added,omitempty"`
	DeletedProperties utils.StringList `json:"deleted,omitempty" yaml:"deleted,omitempty"`
	Diff              *diff.SchemaDiff `json:"changes,omitempty" yaml:"changes,omitempty"`
}

func (in *CRDSchemaDiff) DeepCopy() *CRDSchemaDiff {
	if in == nil {
		return nil
	}

	out := &CRDSchemaDiff{
		AddedProperties:   make(utils.StringList, len(in.AddedProperties)),
		DeletedProperties: make(utils.StringList, len(in.DeletedProperties)),
		Diff:              oasdiff.DeepCopySchemaDiff(in.Diff),
	}

	copy(out.AddedProperties, in.AddedProperties)
	copy(out.DeletedProperties, in.DeletedProperties)

	return out
}

// BreakingChange is, compared to a relatively unspecific Change,
// based on breaking changes reported by oasdiff and tied to
// schema changes.
type BreakingChange struct {
	ID      string        `json:"id" yaml:"id"`
	Level   checker.Level `json:"level" yaml:"level"`
	Details interface{}   `json:"details" yaml:"details"`
}
