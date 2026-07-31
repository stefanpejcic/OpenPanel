import React from "react";
import { DocProvider } from "@docusaurus/plugin-content-docs/client";
import DocItemMetadata from "@theme/DocItem/Metadata";
import DocItemLayout from "@theme/DocItem/Layout";

export const DocItem = (props) => {
    const MDXComponent = props.content;

    return (
        <DocProvider content={props.content}>
            <DocItemMetadata />
            <DocItemLayout>
                <MDXComponent />
            </DocItemLayout>
        </DocProvider>
    );
};
