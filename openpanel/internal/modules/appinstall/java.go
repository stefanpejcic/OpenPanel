package appinstall

import "html/template"

// Java-specific install/run tokens and template-display config, kept apart from nodejs.go/python.go/ruby.go so each type is easy to find and change independently. Unlike the other three, Java is normally compiled, but since Java 11 (JEP 330) `java SomeFile.java` runs a single source file directly with no javac step, so that's the default here, matching the others' "just run one file" model - the optional Maven install step is for pom.xml projects, and running the resulting build still needs a custom run command like `java -jar target/app.jar`.
const javaIconSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#EA2D2E" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-6 h-6 icon icon-tabler icons-tabler-outline icon-tabler-coffee"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M3 14c.83 .642 2.077 1.017 3.5 1c1.423 .017 2.67 -.358 3.5 -1c.83 -.642 2.077 -1.017 3.5 -1c1.423 -.017 2.67 .358 3.5 1" /><path d="M8 3a2.4 2.4 0 0 0 -1 2a2.4 2.4 0 0 0 1 2" /><path d="M12 3a2.4 2.4 0 0 0 -1 2a2.4 2.4 0 0 0 1 2" /><path d="M3 10h14v5a6 6 0 0 1 -6 6h-2a6 6 0 0 1 -6 -6v-5z" /><path d="M16.746 16.726a3 3 0 1 0 -.732 -5.696" /></svg>`

var javaDisplay = kindDisplay{
	Icon:  template.HTML(javaIconSVG), //nolint:gosec // static, server-defined markup, not user input
	Label: "Java", RunCommand: "java", RequiredExtension: ".java",
	RequirementsLabel:   "Run Maven install before starting the app",
	RequirementsTooltip: "When enabled, this option will first run mvn install using the pom.xml file, then launch the application. If the application is already built, you can skip this option.",
}

var Java = Kind{
	AppType: "java", DisplayAppType: "Java", PyOrNode: "JAVA", Title: "Install Java Application",
	InstallToken: "mvn install", RunToken: "java", DefaultStartupFile: "Main.java",
}
