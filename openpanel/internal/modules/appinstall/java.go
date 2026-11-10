package appinstall

import (
	"html/template"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/websites"
)

// Java-specific install/run tokens and template-display config, kept apart from nodejs.go/python.go/ruby.go so each type is easy to find and change independently. Unlike the other three, Java is normally compiled, but since Java 11 (JEP 330) `java SomeFile.java` runs a single source file directly with no javac step, so that's the default here, matching the others' "just run one file" model - the optional Maven install step is for pom.xml projects, and running the resulting build still needs a custom run command like `java -jar target/app.jar`.
const javaIconSVG = `<i class="ri-java-line text-2xl text-gray-500 dark:text-gray-400" aria-hidden="true"></i>`

var javaDisplay = kindDisplay{
	Icon:  template.HTML(javaIconSVG), //nolint:gosec // static, server-defined markup, not user input
	Label: "Java", RunCommand: "java", RequiredExtension: ".java",
	RequirementsLabel:   "Run Maven install before starting the app",
	RequirementsTooltip: "When enabled, this option will first run mvn install using the pom.xml file, then launch the application. If the application is already built, you can skip this option.",
}

var Java = Kind{
	AppType: "java", DisplayAppType: "Java", PyOrNode: "JAVA", Title: "Install Java application",
	InstallToken: "mvn install", InstallCmd: websites.MavenInstallCmd, RunToken: "java", DefaultStartupFile: "Main.java",
}
