package ecs

type InitSystem interface {
	Init(*World)
}

type initSystemManger struct {
	systems []InitSystem
}

func newInitSystemManager() *initSystemManger {
	return &initSystemManger{}
}

func (ism *initSystemManger) registerSystem(s InitSystem) {
	ism.systems = append(ism.systems, s)
}
