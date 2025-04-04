import { Command, Option } from "commander";
import { getProjectType } from "@utils/project";
import { projectScripts } from "../projectScripts";
import { runScript } from "../runScript";
import { updateNotifier } from "src/update-notifier";
import { getPlatformOptionDescription, getRunnerDescription } from "../utils";
import { ProjectTypes } from "@definitions/projectTypes";

const build = (program: Command) => {
    return program
        .command("build")
        .description(getRunnerDescription("build"))
        .allowUnknownOption(true)
        .addOption(
            new Option(
                "-p, --platform <platform>",
                getPlatformOptionDescription(),
            ).choices(
                Object.values(ProjectTypes).filter(
                    (type) => type !== ProjectTypes.UNKNOWN,
                ),
            ),
        )
        .argument("[args...]")
        .action(async (args: string[], { platform }: { platform: ProjectTypes }) => {
            const projectType = getProjectType(platform);
            const binPath = projectScripts[projectType].getBin("build");
            const command = projectScripts[projectType].getBuild(args);

            // Run update notifier only if not already checked
            await updateNotifier();

            // Execute the build script
            await runScript(binPath, command);
        });
};

export default build;
