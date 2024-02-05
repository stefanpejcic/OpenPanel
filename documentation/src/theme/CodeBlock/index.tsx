import React from "react";
import { CodeBlock } from "./base";

export default function CodeBlockWrapper(
    props: JSX.IntrinsicAttributes & {
        live?: boolean;
        shared?: boolean;
        className?: string;
    },
): JSX.Element {
    return <CodeBlock {...(props as any)} />;
}
