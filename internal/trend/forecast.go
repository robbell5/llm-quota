package trend

import (
	"fmt"
	"time"
)

// Forecast describes the rendered pace signal for one window.
type Forecast struct {
	Rate   float64 // percent per hour, >= 0
	AtRisk bool    // projected to hit 100% before reset
	Arrow  rune    // '↑' at-risk, '→' otherwise
	Status string  // "", "~91% by reset", "full in 40m", or "full"
}

// ComputeForecast derives the forecast from the current reading and burn rate.
func ComputeForecast(usedPct, ratePerHour float64, now, resetsAt time.Time) Forecast {
	f := Forecast{Rate: ratePerHour, Arrow: '→'}

	if usedPct >= 100 {
		f.Status = "full"
		return f
	}

	hoursUntilReset := resetsAt.Sub(now).Hours()
	if hoursUntilReset <= 0 {
		return f // window resetting; show rate only
	}

	if ratePerHour <= 0 {
		f.Status = fmt.Sprintf("~%.0f%% by reset", usedPct)
		return f
	}

	hoursToFull := (100 - usedPct) / ratePerHour
	if hoursToFull < hoursUntilReset {
		f.AtRisk = true
		f.Arrow = '↑'
		f.Status = "full in " + shortDuration(hoursToFull)
		return f
	}

	proj := usedPct + ratePerHour*hoursUntilReset
	if proj > 100 {
		proj = 100
	}
	f.Status = fmt.Sprintf("~%.0f%% by reset", proj)
	return f
}

func shortDuration(hours float64) string {
	d := time.Duration(hours * float64(time.Hour))
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d/time.Minute))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d/time.Hour), int((d%time.Hour)/time.Minute))
	}
	return fmt.Sprintf("%dd", int(d/(24*time.Hour)))
}
