import React, { useMemo } from "react";
import Link from "@docusaurus/Link";
import { useDocsSidebar, useDocsVersion } from "@docusaurus/plugin-content-docs/client";

type Item = { type: string; label?: string; href?: string; docId?: string; items?: Item[] };

// hand picked, newest first; the version pages hold the full lists
export const HIGHLIGHTS: [string, string, string][] = [
    ["2.0.32", "Golub webmail", "A fast, privacy-first webmail you can switch to from Roundcube, with the same one-click login from OpenPanel."],
    ["2.0.28", "OpenCLI rewritten in Go", "Every opencli command in one binary, about 2x faster with the same commands and output."],
    ["2.0.28", "PHP 8.6 and user panel themes", "PHP 8.6 by default for new users, Saturn and Modern themes, and users can pick their own theme."],
    ["2.0.25", "32 new one-click Docker apps", "Open WebUI, Langflow, Excalidraw, Stirling PDF and more in the Auto Installer."],
    ["2.0.24", "ImunifyAV for every user", "Users get ImunifyAV's own interface and a summary of infected files and scans."],
    ["2.0.23", "WHMCS reseller accounts", "The WHMCS module can sell reseller accounts, with new reseller API endpoints."],
];

const link = "text-refine-react-light-link dark:text-refine-react-dark-link hover:underline";

export default function ChangelogOverview() {
    const sidebar = useDocsSidebar();
    const version = useDocsVersion();

    const { latest, upcoming, count, last90 } = useMemo(() => {
        const docs = version.docs as Record<string, any>;
        const versions = ((sidebar?.items ?? []) as Item[])
            .filter((i) => i.type === "link" && i.docId && /^\d+\.\d+\.\d+$/.test(i.label ?? ""))
            .map((i) => {
                // the first line of each version page is "Released on <date>", unreleased ones don't have it yet
                const m = /^Released on ([A-Za-z]+ \d{1,2}, \d{4})/.exec(docs[i.docId!]?.description ?? "");
                return { v: i.label!, href: i.href!, date: m ? m[1] : undefined };
            });
        const released = versions.filter((x) => x.date);
        const cutoff = Date.now() - 90 * 864e5;
        return {
            latest: released[0],
            upcoming: versions.find((x) => !x.date),
            count: released.length,
            last90: released.filter((x) => Date.parse(x.date!) >= cutoff).length,
        };
    }, [sidebar, version]);

    const card = "rounded-2xl border border-gray-300 dark:border-gray-700 p-5 flex flex-col gap-1 no-underline";

    return (
        <div className="not-prose flex flex-col gap-8 mt-2 mb-10">
            <div className="grid gap-3 sm:grid-cols-3">
                {latest && (
                    <Link to={latest.href} className={`${card} hover:border-gray-400 dark:hover:border-gray-500`}>
                        <span className="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">Latest release</span>
                        <span className="text-2xl font-bold text-gray-900 dark:text-gray-0">{latest.v}</span>
                        <span className="text-sm text-gray-600 dark:text-gray-400">{latest.date}</span>
                    </Link>
                )}
                {upcoming && (
                    <Link to={upcoming.href} className={`${card} hover:border-gray-400 dark:hover:border-gray-500`}>
                        <span className="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">In progress</span>
                        <span className="text-2xl font-bold text-gray-900 dark:text-gray-0">{upcoming.v}</span>
                        <span className="text-sm text-gray-600 dark:text-gray-400">Fixes and changes merged so far</span>
                    </Link>
                )}
                <div className={card}>
                    <span className="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">Release pace</span>
                    <span className="text-2xl font-bold text-gray-900 dark:text-gray-0">{last90} releases</span>
                    <span className="text-sm text-gray-600 dark:text-gray-400">in the last 90 days, {count} since 1.0.0</span>
                </div>
            </div>

            <section className="flex flex-col gap-4">
                <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Recent highlights</h2>
                <ol className="m-0 p-0 list-none flex flex-col">
                    {HIGHLIGHTS.map(([v, title, desc], i) => (
                        <li key={title} className="flex gap-4">
                            <div className="flex flex-col items-center">
                                <span className="mt-1.5 w-2.5 h-2.5 shrink-0 rounded-full bg-refine-blue dark:bg-refine-cyan-alt" />
                                {i < HIGHLIGHTS.length - 1 && <span className="w-px flex-1 bg-gray-300 dark:bg-gray-700" />}
                            </div>
                            <div className="flex flex-col gap-0.5 pb-5">
                                <div className="flex flex-wrap items-baseline gap-2">
                                    <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{title}</span>
                                    <Link to={`/docs/changelog/${v}/`} className={`text-xs font-mono ${link}`}>
                                        {v}
                                    </Link>
                                </div>
                                <span className="text-sm text-gray-600 dark:text-gray-400">{desc}</span>
                            </div>
                        </li>
                    ))}
                </ol>
            </section>

            <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                Updates install automatically when auto-updates are on, or manually with{" "}
                <code className="text-xs">opencli update --force</code>. See what's planned on the{" "}
                <a href="/roadmap/" className={link}>
                    roadmap
                </a>{" "}
                or follow{" "}
                <a href="https://github.com/stefanpejcic/OpenPanel/milestones" className={link}>
                    milestones on GitHub
                </a>
                .
            </p>
        </div>
    );
}
