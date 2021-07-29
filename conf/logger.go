package conf

// LoggerConfig log configs object
type LoggerConfig struct {
	Level      int
	Path       string
	Format     string
	Output     string
	OutputFile string
}

func getLogConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:      c.GetInt("logger.level"),
		Format:     c.GetString("logger.format"),
		Path:       c.GetString("logger.path"),
		Output:     c.GetString("logger.output"),
		OutputFile: c.GetString("logger.output_file"),
	}
}
