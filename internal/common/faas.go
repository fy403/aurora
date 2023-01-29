package common

import "aurora/internal/config"

type Faas struct {
	cnf *config.Faas
}

func NewFaas(cnf *config.Faas) Faas {
	return Faas{cnf: cnf}
}

func (f *Faas) GetConfig() *config.Faas {
	return f.cnf
}
