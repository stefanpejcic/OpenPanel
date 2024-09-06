import React, { useEffect, useState } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";

interface Milestone {
    id: number;
    title: string;
    description: string;
    due_on: string;
    open_issues: number;
    closed_issues: number;
    html_url: string;
}

const Translate: React.FC = () => {
    const [readmeContent, setReadmeContent] = useState<string | null>(null);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchReadme = async () => {
            try {
                const response = await fetch(
                    "https://raw.githubusercontent.com/stefanpejcic/openpanel-translations/main/README.md"
                );
                if (!response.ok) {
                    throw new Error("Failed to fetch README.md content.");
                }
                const data = await response.text();
                setReadmeContent(data);
            } catch (err) {
                setError((err as Error).message);
            } finally {
                setLoading(false);
            }
        };

        fetchReadme();
    }, []);

    return (
        <CommonLayout>
            <Head title="Translations | OpenPanel">
                <html data-page="translate" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    <h1>Translations</h1>

                    {/* Loading State */}
                    {loading && <p>Loading translations data...</p>}

                    {/* Error State */}
                    {error && <p className="text-red-500">Error: {error}</p>}

                    {/* Display README Content */}
                    {readmeContent && (
                        <div className="mt-4">
                            <pre
                                style={{
                                    whiteSpace: "pre-wrap",
                                    background: "#f6f8fa",
                                    padding: "16px",
                                    borderRadius: "8px",
                                }}
                            >
                                {readmeContent}
                            </pre>
                        </div>
                    )}
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Translate;
