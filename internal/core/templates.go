package core

import (
	"fmt"
	"os"
	"path/filepath"

	"poly-switch/internal/config"
)

var defaultTemplates = map[string]string{
	"maven-aliyun.xml": `<?xml version="1.0" encoding="UTF-8"?>
<settings xmlns="http://maven.apache.org/SETTINGS/1.0.0"
          xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
          xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.0.0 http://maven.apache.org/xsd/settings-1.0.0.xsd">
  <mirrors>
    <mirror>
      <id>aliyunmaven</id>
      <mirrorOf>*</mirrorOf>
      <name>阿里云公共仓库</name>
      <url>https://maven.aliyun.com/repository/public</url>
    </mirror>
  </mirrors>
</settings>
`,
	"gradle-aliyun.gradle": `allprojects {
    repositories {
        maven { url 'https://maven.aliyun.com/repository/public/' }
        mavenLocal()
        mavenCentral()
    }
}
`,
}

// EnsureTemplates writes the default mirror templates to the user's config directory.
func EnsureTemplates() error {
	polyDir, err := config.PolySwitchDir()
	if err != nil {
		return err
	}
	templatesDir := filepath.Join(polyDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return err
	}

	for name, content := range defaultTemplates {
		path := filepath.Join(templatesDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return fmt.Errorf("write template %s: %w", name, err)
			}
		}
	}
	return nil
}
