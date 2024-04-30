import React, { useEffect, useState } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";
import marked from "marked"; // Import the marked library for parsing Markdown

const Translations: React.FC = () => {
    const [readmeContent, setReadmeContent] = useState<string>("");

    useEffect(() => {
        const fetchReadme = async () => {
            try {
                const response = await fetch("https://raw.githubusercontent.com/stefanpejcic/openpanel-translations/main/README.md");
                const data = await response.text();
                setReadmeContent(data);
            } catch (error) {
                console.error("Error fetching README:", error);
            }
        };

        fetchReadme();
    }, []);

    const renderMarkdown = (content: string) => {
        // Use marked library to parse Markdown content
        return {__html: marked(content)};
    };

    return (
        <CommonLayout>

        </CommonLayout>
    );
};

export default Translations;
