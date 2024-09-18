import React from "react";
import { CodeBlock } from "../theme/CodeBlock/base";
import { CommandLineIcon } from "./icons/command-line";
import { TerminalIcon } from "./icons/terminal";

type Props = {
    path: string;
};

export const CommonRunLocalPrompt = ({ path }: Props) => {
    return (
        <CodeBlock
            language="bash"
            title="Install command"
            icon={<TerminalIcon />}
        >
            {`bash <(curl -sSL https://openpanel.org) ${path}`}
        </CodeBlock>
    );
};
