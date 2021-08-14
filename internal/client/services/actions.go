package services

type action int

const (
	create action = iota
	cancel
	statistic
)

type actionPlane struct {
	action    action
	ptr       int
	questions []string
	answers   []string
}

func createPlan() actionPlane {
	plan := actionPlane{
		action: create,
		questions: []string{
			"Введите название тикера",
			"Введите тип операции (BUY/SELL)",
			"Введите желаемое количество",
			"Введите желаемую стоимость",
		},
	}

	plan.answers = make([]string, len(plan.questions))

	return plan
}

func cancelPlan() actionPlane {
	plan := actionPlane{
		action: cancel,
		questions: []string{
			"Введите идентификатор сделки",
		},
	}

	plan.answers = make([]string, len(plan.questions))

	return plan
}

func statPlan() actionPlane {
	plan := actionPlane{
		action: statistic,
		questions: []string{
			"Введите название тикера",
		},
	}

	plan.answers = make([]string, len(plan.questions))

	return plan
}
