package reporter

import (
	"fmt"
	"strings"

	"github.com/younsl/ghes-schedule-scanner/pkg/models"
)

type ReportFormatter interface {
	FormatReport(*models.ScanResult) string
}

type ConsoleFormatter struct{}

type Reporter struct {
	formatter ReportFormatter
}

func NewReporter(formatter ReportFormatter) *Reporter {
	return &Reporter{formatter: formatter}
}

func (r *Reporter) GenerateReport(result *models.ScanResult) error {
	output := r.formatter.FormatReport(result)
	fmt.Print(output)
	return nil
}

func (f *ConsoleFormatter) FormatReport(result *models.ScanResult) string {
	if result == nil || len(result.Workflows) == 0 {
		return "No workflows found\n"
	}

	var sb strings.Builder
	sb.WriteString("Scheduled Workflows Summary:\n")
	sb.WriteString(fmt.Sprintf("%-3s   %-35s   %-35s   %-13s   %-13s   %s\n",
		"NO", "REPOSITORY", "WORKFLOW", "UTC SCHEDULE", "KST SCHEDULE", "LAST STATUS"))

	for i, wf := range result.Workflows {
		for _, schedule := range wf.CronSchedules {
			sb.WriteString(fmt.Sprintf("%-3d   %-35s   %-35s   %-13s   %-13s   %s\n",
				i+1,
				truncateString(wf.RepoName, 35),
				truncateString(wf.WorkflowName, 35),
				schedule,
				convertToKST(schedule),
				wf.LastStatus,
			))
		}
	}

	sb.WriteString(fmt.Sprintf("\nScanned: %d repos, %d workflows\n",
		result.TotalRepos,
		len(result.Workflows)))
	return sb.String()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s + strings.Repeat(" ", maxLen-len(s))
	}
	return s[:maxLen-2] + ".."
}

func convertToKST(utcCron string) string {
	parts := strings.Split(utcCron, " ")
	if len(parts) != 5 {
		return utcCron
	}

	if !isValidCronExpression(parts) {
		return utcCron
	}

	minute := parts[0]
	hour := atoi(parts[1])
	dom := parts[2]
	month := parts[3]
	dow := parts[4]

	newHour := (hour + 9) % 24
	dayShift := (hour + 9) / 24

	if dayShift > 0 && dow != "*" {
		dowNum := atoi(dow)
		if dowNum >= 0 && dowNum <= 6 {
			newDow := (dowNum + 1) % 7
			dow = fmt.Sprintf("%d", newDow)
		}
	}

	return fmt.Sprintf("%s %d %s %s %s", minute, newHour, dom, month, dow)
}

func isValidCronExpression(parts []string) bool {
	if parts[0] != "*" && (atoi(parts[0]) < 0 || atoi(parts[0]) > 59) {
		return false
	}
	if parts[1] != "*" && (atoi(parts[1]) < 0 || atoi(parts[1]) > 23) {
		return false
	}
	if parts[4] != "*" && (atoi(parts[4]) < 0 || atoi(parts[4]) > 6) {
		return false
	}
	return true
}

func atoi(s string) int {
	if s == "*" {
		return 0
	}
	i := 0
	if _, err := fmt.Sscanf(s, "%d", &i); err != nil {
		return 0
	}
	return i
}