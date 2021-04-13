package runner

type Runner interface {
	Start()
	Stop()
}

type Registrator interface {
	Register(r Runner)
}
