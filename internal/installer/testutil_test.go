package installer

// MockFormatter is a test implementation of OutputFormatter
type MockFormatter struct {
	Sections  []string
	Tasks     []string
	Successes []string
	Errors    []string
	Warnings  []string
	Progress  []string
}

func (m *MockFormatter) PrintSection(sectionName string) {
	m.Sections = append(m.Sections, sectionName)
}

func (m *MockFormatter) PrintProgress(name, output string) {
	m.Progress = append(m.Progress, name+": "+output)
}

func (m *MockFormatter) PrintSuccess(name, output string) {
	m.Successes = append(m.Successes, name+": "+output)
}

func (m *MockFormatter) PrintError(name, output string) {
	m.Errors = append(m.Errors, name+": "+output)
}

func (m *MockFormatter) PrintWarning(name, output string) {
	m.Warnings = append(m.Warnings, name+": "+output)
}

func (m *MockFormatter) PrintTask(section, name, output string) {
	m.Tasks = append(m.Tasks, section+": "+name+": "+output)
}
