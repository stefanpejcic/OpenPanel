import React from "react";
import clsx from "clsx";

type PanelKey = "openpanel" | "cpanel" | "plesk" | "directadmin" | "cyberpanel" | "enhance";

// minimum RAM from each vendor's own install requirements, checked November 2026
const PANELS: { key: PanelKey; name: string; ramGb: number; note?: string; source: string }[] = [
    { key: "openpanel", name: "OpenPanel", ramGb: 1, source: "/docs/articles/install-update/system-requirements/" },
    { key: "plesk", name: "Plesk", ramGb: 1, note: "plus 1 GB swap", source: "https://docs.plesk.com/release-notes/obsidian/hardware-requirements/" },
    { key: "cyberpanel", name: "CyberPanel", ramGb: 1, source: "https://cyberpanel.net/KnowledgeBase/home/install-cyberpanel/" },
    { key: "cpanel", name: "cPanel & WHM", ramGb: 2, note: "4 GB recommended", source: "https://docs.cpanel.net/installation-guide/system-requirements/" },
    { key: "enhance", name: "Enhance", ramGb: 2, source: "https://enhance.com/docs/getting-started/requirements" },
    { key: "directadmin", name: "DirectAdmin", ramGb: 4, note: "plus 4 GB swap", source: "https://docs.directadmin.com/getting-started/installation/overview.html" },
];

const MAX = Math.max(...PANELS.map((p) => p.ramGb));

export const ResourceComparison = ({ competitor, className }: { competitor: PanelKey; className?: string }) => {
    const rival = PANELS.find((p) => p.key === competitor)!;
    return (
        <section className={clsx("not-prose px-4 landing-sm:px-6 landing-md:px-10", className)}>
            <div className="rounded-2xl border border-gray-300 dark:border-gray-700 p-5 sm:p-8 grid gap-8 landing-md:grid-cols-[1fr_1.4fr]">
                <div className="flex flex-col gap-4">
                    <h2 className="text-2xl landing-sm:text-[32px] landing-sm:leading-[40px] font-semibold text-gray-900 dark:text-gray-0 m-0">
                        Resource usage: OpenPanel vs {rival.name}
                    </h2>
                    <p className="text-base text-gray-600 dark:text-gray-400 m-0">
                        OpenPanel uses about <b className="text-gray-900 dark:text-gray-0">53 MB of RAM</b> at idle (user panel
                        ~20 MB, admin panel ~33 MB) and about 324 MB of disk, so almost all of the server's memory
                        goes to your users' websites.
                    </p>
                    <div className="grid grid-cols-2 gap-3">
                        <div className="rounded-xl bg-gray-50 dark:bg-gray-800/60 p-4">
                            <div className="text-2xl font-bold text-gray-900 dark:text-gray-0">~53 MB</div>
                            <div className="text-sm text-gray-600 dark:text-gray-400">panel RAM at idle</div>
                        </div>
                        <div className="rounded-xl bg-gray-50 dark:bg-gray-800/60 p-4">
                            <div className="text-2xl font-bold text-gray-900 dark:text-gray-0">~15 MB</div>
                            <div className="text-sm text-gray-600 dark:text-gray-400">RAM per idle account</div>
                        </div>
                    </div>
                    <a
                        href="/calculator/"
                        className="self-start text-sm text-refine-react-light-link dark:text-refine-react-dark-link hover:underline"
                    >
                        Size your server with the resource calculator →
                    </a>
                </div>

                <div className="flex flex-col gap-3">
                    <span className="text-sm font-semibold text-gray-700 dark:text-gray-300">Minimum RAM to install</span>
                    <ul className="m-0 p-0 list-none flex flex-col gap-2.5">
                        {PANELS.map((p) => {
                            const ours = p.key === "openpanel";
                            const focus = ours || p.key === competitor;
                            return (
                                <li key={p.key} className={clsx("flex flex-col gap-1", !focus && "opacity-60")}>
                                    <div className="flex items-baseline justify-between gap-3 text-sm">
                                        <a
                                            href={p.source}
                                            className={clsx(
                                                "no-underline hover:underline",
                                                focus ? "font-semibold text-gray-900 dark:text-gray-100" : "text-gray-700 dark:text-gray-300",
                                            )}
                                        >
                                            {p.name}
                                        </a>
                                        <span className="font-mono text-gray-900 dark:text-gray-100">
                                            {p.ramGb} GB
                                            {p.note && <span className="font-sans text-xs text-gray-500 dark:text-gray-400"> · {p.note}</span>}
                                        </span>
                                    </div>
                                    <div className="h-2.5 rounded-full bg-gray-100 dark:bg-gray-800 overflow-hidden">
                                        <div
                                            className={clsx(
                                                "h-full rounded-full",
                                                ours ? "bg-refine-blue dark:bg-refine-cyan-alt" : p.key === competitor ? "bg-gray-500" : "bg-gray-300 dark:bg-gray-600",
                                            )}
                                            style={{ width: `${(p.ramGb / MAX) * 100}%` }}
                                        />
                                    </div>
                                </li>
                            );
                        })}
                    </ul>
                    <p className="text-xs text-gray-500 dark:text-gray-400 m-0">
                        From each vendor's published install requirements, click a name for the source. Real usage depends on
                        your sites and traffic.
                    </p>
                </div>
            </div>
        </section>
    );
};
