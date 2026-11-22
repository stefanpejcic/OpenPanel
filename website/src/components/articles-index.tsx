import React, { useMemo, useState } from "react";
import clsx from "clsx";
import Link from "@docusaurus/Link";
import { useDocsSidebar, useDocsVersion } from "@docusaurus/plugin-content-docs/client";

type Item = { type: string; label?: string; href?: string; docId?: string; items?: Item[] };
type Article = { title: string; href: string; description: string; category: string };

const START_HERE = [
    ["How to access OpenAdmin", "/docs/articles/dev-experience/how-to-access-openadmin/", "Log in to the admin panel after the install."],
    ["How to access OpenPanel", "/docs/articles/dev-experience/how-to-access-openpanel/", "Where your users log in to manage their sites."],
    ["Frequently asked questions", "/docs/articles/faq/", "Pricing, requirements, supported systems and more."],
    ["Start a web hosting business", "/docs/articles/hosting-business/start-web-hosting-business/", "Plans, billing and everything to sell hosting."],
    ["Import a cPanel backup", "/docs/articles/transfers/import-cpanel-backup-to-openpanel/", "Move accounts over from cPanel."],
    ["Try the live demo", "/docs/articles/support/how-to-access-openpanel-demo/", "Click around OpenPanel and OpenAdmin first."],
];

const POPULAR = [
    ["403 error", "/docs/articles/domains/403-error-on-website-disable-coraza-waf/"],
    ["502 error", "/docs/articles/domains/bad-gateway-502-error-troubleshooting"],
    ["504 error", "/docs/articles/domains/gateway-timeout-504-error-troubleshooting"],
    ["Install WordPress", "/docs/articles/websites/how-to-install-wordpress-with-openpanel/"],
    ["PHP memory_limit", "/docs/articles/websites/how-to-set-or-increase-PHP-INI-memory-limit-or-other-values/"],
    ["Emails go to spam", "/docs/articles/email/why-emails-go-to-spam/"],
    ["Update manually", "/docs/articles/install-update/how-to-manually-update-openpanel/"],
];

const link = "text-refine-react-light-link dark:text-refine-react-dark-link hover:underline";
const SHOWN = 4;

// every doc under a sidebar item, nested categories included
const collect = (item: Item, docs: Record<string, any>, category: string, out: Article[]) => {
    // in the client a sidebar doc is a "link" item with a docId
    if (item.type === "link" && item.docId) {
        const d = item.docId ? docs[item.docId] : undefined;
        if (item.href && !/\/intro\/?$/.test(item.href))
            out.push({ title: d?.title ?? item.label ?? "", href: item.href, description: d?.description ?? "", category });
    }
    for (const child of item.items ?? []) collect(child, docs, category, out);
};

export default function ArticlesIndex() {
    const sidebar = useDocsSidebar();
    const version = useDocsVersion();
    const [query, setQuery] = useState("");
    const [open, setOpen] = useState<Record<string, boolean>>({});

    const categories = useMemo(() => {
        const docs = version.docs as Record<string, any>;
        return ((sidebar?.items ?? []) as Item[])
            .filter((i) => i.type === "category")
            .map((c) => {
                const articles: Article[] = [];
                for (const child of c.items ?? []) collect(child, docs, c.label ?? "", articles);
                return { label: c.label ?? "", href: c.href, articles };
            })
            .filter((c) => c.articles.length);
    }, [sidebar, version]);

    const all = useMemo(() => categories.flatMap((c) => c.articles), [categories]);
    const q = query.trim().toLowerCase();
    const results = q
        ? all.filter((a) => a.title.toLowerCase().includes(q) || a.description.toLowerCase().includes(q))
        : [];

    return (
        <div className="not-prose flex flex-col gap-10 mt-2">
            <div className="flex flex-col gap-3">
                <input
                    type="search"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    placeholder={`Search ${all.length} guides, e.g. backup, DNS, WordPress`}
                    aria-label="Search guides"
                    className="w-full rounded-lg border border-gray-300 dark:border-gray-700 bg-gray-0 dark:bg-gray-900 text-gray-900 dark:text-gray-100 placeholder:text-gray-400 px-4 py-3 text-base outline-none focus:ring-2 focus:ring-refine-blue/40"
                />
                {!q && (
                    <div className="flex flex-wrap items-center gap-2 text-sm">
                        <span className="text-gray-500 dark:text-gray-400">Popular:</span>
                        {POPULAR.map(([label, href]) => (
                            <Link
                                key={href}
                                to={href}
                                className="rounded-full px-3 py-1 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 hover:border-gray-400 no-underline"
                            >
                                {label}
                            </Link>
                        ))}
                    </div>
                )}
            </div>

            {q ? (
                <section className="flex flex-col gap-3">
                    <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                        {results.length} {results.length === 1 ? "guide" : "guides"}
                    </p>
                    <ul className="m-0 p-0 list-none flex flex-col divide-y divide-gray-200 dark:divide-gray-700 border border-gray-200 dark:border-gray-700 rounded-xl">
                        {results.map((a) => (
                            <li key={a.href} className="p-4">
                                <Link to={a.href} className="no-underline flex flex-col gap-1">
                                    <span className="text-xs text-gray-500 dark:text-gray-400">{a.category}</span>
                                    <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{a.title}</span>
                                    {a.description && (
                                        <span className="text-sm text-gray-600 dark:text-gray-400 line-clamp-2">{a.description}</span>
                                    )}
                                </Link>
                            </li>
                        ))}
                    </ul>
                </section>
            ) : (
                <>
                    <section className="flex flex-col gap-4">
                        <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Start here</h2>
                        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                            {START_HERE.map(([title, href, desc]) => (
                                <Link
                                    key={href}
                                    to={href}
                                    className="rounded-xl border border-gray-200 dark:border-gray-700 p-4 flex flex-col gap-1 no-underline hover:border-gray-400 dark:hover:border-gray-500"
                                >
                                    <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{title} →</span>
                                    <span className="text-sm text-gray-600 dark:text-gray-400">{desc}</span>
                                </Link>
                            ))}
                        </div>
                    </section>

                    <section className="flex flex-col gap-4">
                        <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">All guides by topic</h2>
                        <div className="grid gap-4 sm:grid-cols-2">
                            {categories.map((c) => {
                                const expanded = open[c.label];
                                const list = expanded ? c.articles : c.articles.slice(0, SHOWN);
                                return (
                                    <div key={c.label} className="rounded-2xl border border-gray-300 dark:border-gray-700 p-5 flex flex-col gap-3">
                                        <div className="flex items-baseline justify-between gap-3">
                                            <h3 className="text-base font-semibold text-gray-900 dark:text-gray-0 m-0">
                                                {c.href ? (
                                                    <Link to={c.href} className="no-underline text-gray-900 dark:text-gray-0 hover:underline">
                                                        {c.label}
                                                    </Link>
                                                ) : (
                                                    c.label
                                                )}
                                            </h3>
                                            <span className="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">
                                                {c.articles.length} {c.articles.length === 1 ? "guide" : "guides"}
                                            </span>
                                        </div>
                                        <ul className="m-0 p-0 list-none flex flex-col gap-1.5">
                                            {list.map((a) => (
                                                <li key={a.href} className="text-sm">
                                                    <Link to={a.href} className={link}>
                                                        {a.title}
                                                    </Link>
                                                </li>
                                            ))}
                                        </ul>
                                        {c.articles.length > SHOWN && (
                                            <button
                                                type="button"
                                                onClick={() => setOpen((s) => ({ ...s, [c.label]: !expanded }))}
                                                className={clsx("self-start text-sm font-semibold bg-transparent border-0 p-0 cursor-pointer", link)}
                                            >
                                                {expanded ? "Show less" : `Show all ${c.articles.length}`}
                                            </button>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </section>
                </>
            )}

            <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                Can't find what you need? Ask on{" "}
                <a href="https://github.com/stefanpejcic/OpenPanel/discussions" className={link}>
                    GitHub Discussions
                </a>{" "}
                or{" "}
                <a href="https://discord.openpanel.org/" className={link}>
                    Discord
                </a>
                .
            </p>
        </div>
    );
}
