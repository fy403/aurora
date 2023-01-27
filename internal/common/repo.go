package common

import "aurora/internal/config"

type Repo struct {
	cnf *config.Config
}

func NewRepo(cnf *config.Config) Repo {
	return Repo{cnf: cnf}
}

func (r *Repo) GetConfig() *config.Config {
	return r.cnf
}
