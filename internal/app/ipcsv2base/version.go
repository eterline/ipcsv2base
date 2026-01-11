// Copyright (c) 2026 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package ipcsv2base

import "github.com/eterline/ipcsv2base/internal/model"

type InitFlags struct {
	CommitHash string
	Version    string
	Repository string
}

func (inf InitFlags) IsDev() bool {
	return inf.GetCommitHash() == "dev" || inf.GetVersion() == "dev"
}

func (inf InitFlags) GetCommitHash() string {
	return inf.CommitHash
}

func (inf InitFlags) GetVersion() string {
	return inf.Version
}

func (inf InitFlags) GetRepository() string {
	return inf.Repository
}

func (inf InitFlags) FieldsLog() []model.LogField {
	return []model.LogField{
		model.FieldString("commit", inf.GetCommitHash()),
		model.FieldString("version", inf.GetVersion()),
		model.FieldString("repository", inf.GetRepository()),
	}
}
