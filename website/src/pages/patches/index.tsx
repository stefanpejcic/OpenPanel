import React, { useEffect, useState } from "react";
import Head from "@docusaurus/Head";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";

interface Patch {
    name: string;
    date: string;
    description: string;
    url: string;
}

const GITHUB_PATCHES_API =
    "https://api.github.com/repos/stefanpejcic/OpenPanel/contents/patches";

const Patches: React.FC = () => {
    const [patches, setPatches] = useState<Patch[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchPatches = async () => {
            try {
                const response = await fetch(GITHUB_PATCHES_API);
                const files = await response.json();

                // Filter only `.sh` files
                const shFiles = files.filter((file: any) => file.name.endsWith(".sh"));

                const patchPromises = shFiles.map(async (file: any) => {
                    const rawFile = await fetch(file.download_url);
                    const content = await rawFile.text();

                    const descMatch = content.match(/#\s*Description:\s*([\s\S]*?)(?=\n#|$)/i);
                    const dateMatch = content.match(/#\s*Date:\s*(.*)/i);

                    return {
                        name: file.name.replace(/\.sh$/, ""),
                        date: dateMatch ? dateMatch[1].trim() : "N/A",
                        description: descMatch
                            ? descMatch[1].replace(/\n#\s*/g, " ").trim()
                            : "No description found.",
                        url: file.html_url,
                    };
                });

                const results = await Promise.all(patchPromises);
                setPatches(results);
            } catch (error) {
                console.error("Error loading patches:", error);
            } finally {
                setLoading(false);
            }
        };

        fetchPatches();
    }, []);

    return (
        <CommonLayout>
            <Head title="OpenPanel Patches | OpenPanel">
                <html data-page="patches" data-customized="true" />
            </Head>

            <div className="refine-prose">
                <CommonHeader hasSticky={true} />
                <div className="flex flex-col items-center pt-8 lg:pt-16 pb-32 max-w-[900px] w-full mx-auto px-4">
                    <h1>OpenPanel Patches</h1>
                    <p className="text-gray-500 dark:text-gray-400 text-base mb-6 text-center">
                        No product is perfect. Bugs are going to happen in a product that evolves over time.
                        As soon as a bug is reported on Github issues, we do our best to resolve it immediately and include the fix in the next update.
                        If you do not wish to wait for that update and want to apply the fix immediately, you can install a patch:
                    </p>

                    <h2 className="mt-10 mb-4">Usage</h2>
                    <pre className="bg-gray-900 text-white rounded-xl p-4 w-full overflow-x-auto">
                        opencli patch &lt;PATCH_NAME&gt;
                    </pre>

                    <h2 className="mt-10 mb-6">Available Patches</h2>

                    {loading ? (
                        <p>Loading patches...</p>
                    ) : patches.length === 0 ? (
                        <p>No patches found, please check manually in: https://api.github.com/repos/stefanpejcic/OpenPanel/contents/patches </p>
                    ) : (
                        <table className="w-full border border-gray-200 dark:border-gray-800 text-left rounded-xl overflow-hidden">
                            <thead className="bg-gray-100 dark:bg-gray-800">
                                <tr>
                                    <th className="py-3 px-4 font-semibold">Name</th>
                                    <th className="py-3 px-4 font-semibold">Date</th>
                                    <th className="py-3 px-4 font-semibold">Description</th>
                                </tr>
                            </thead>
                            <tbody>
                                {patches.map((patch) => (
                                    <tr
                                        key={patch.name}
                                        className="border-t border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-900"
                                    >
                                        <td className="py-3 px-4 font-mono">
                                            <a
                                                href={patch.url}
                                                target="_blank"
                                                rel="noopener noreferrer"
                                                className="text-blue-600 dark:text-blue-400 hover:underline"
                                            >
                                                {patch.name}
                                            </a>
                                        </td>
                                        <td className="py-3 px-4 text-sm text-gray-600 dark:text-gray-400">
                                            {patch.date}
                                        </td>
                                        <td className="py-3 px-4 text-sm text-gray-700 dark:text-gray-300">
                                            {patch.description}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    )}
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Patches;
                                                  
