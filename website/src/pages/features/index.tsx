import React, { useEffect, useMemo, useState } from "react";
import clsx from "clsx";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { integrations as data } from "../../assets/integrations";

type Edition = "all" | "community" | "enterprise";

interface Category {
    id: string;
    title: string;
    lead: string;
    points: string[];
    screenshot?: string;
    alt?: string;
    // false when there's no _dark.png next to the screenshot
    dark?: boolean;
    // names from assets/integrations.ts, icons and descriptions come from there
    items: string[];
}

const CATEGORIES: Category[] = [
    {
        id: "isolation",
        title: "Isolation & security",
        lead: "Every account runs in its own containers, so one site can't see, slow down or break another.",
        points: [
            "Own web server, database and PHP per account",
            "CPU, RAM, I/O and process limits per plan and per service",
            "Firewall, WAF and malware scanning built in",
        ],
        screenshot: "/img/openpanel-screenshots/containers/containers-list.png",
        alt: "Services of one account running in their own containers",
        items: [
            "Containerized services",
            "Resource limiting",
            "Resource usage management",
            "Webserver per user",
            "Database server per user",
            "Switch web servers",
            "Firewall",
            "CorazaWAF",
            "ImunifyAV",
            "ClamAV Malware Scanner",
            "IP Blocker",
            "Two-Factor Authentication",
            "Passkeys",
            "Login History",
            "Activity logs",
            "Disable admin panel",
            "Separate services",
        ],
    },
    {
        id: "websites",
        title: "Websites & apps",
        lead: "Pick the web server, PHP version and stack per account, then install WordPress or an app in one click.",
        points: [
            "Apache, Nginx, OpenResty, LiteSpeed or OpenLiteSpeed",
            "Several PHP versions side by side, per domain",
            "WordPress manager and a one-click app installer",
        ],
        screenshot: "/img/openpanel-screenshots/applications/autoinstaller-page.png",
        alt: "One-click app installer",
        items: [
            "Apache",
            "Nginx",
            "OpenResty",
            "LiteSpeed",
            "OpenLiteSpeed",
            "Domain Names",
            "SSL certificates",
            "PHP versions",
            "WP Manager",
            "One-Click App Installer",
            "NodeJS, Python, Ruby & Java",
            "Varnish",
            "Object Caching",
            "Server settings",
            "Cron Jobs",
            "DNS Zone Editor",
            "Dynamic DNS",
            "Visitor reports",
        ],
    },
    {
        id: "email-databases",
        title: "Email & databases",
        lead: "Mailboxes, webmail and spam filtering next to MySQL, PostgreSQL and MongoDB, all managed from the same panel.",
        points: [
            "Email accounts, aliases, forwarders and filters",
            "MySQL, MariaDB or Percona with phpMyAdmin auto-login",
            "PostgreSQL and MongoDB when users need them",
        ],
        screenshot: "/img/openpanel-screenshots/emails/emails-list.png",
        alt: "Email accounts list",
        items: ["Email Hosting", "MySQL and phpMyAdmin", "PostgreSQL", "MongoDB"],
    },
    {
        id: "files",
        title: "Files & backups",
        lead: "Users manage their own files and backups, and accounts move over from cPanel in one step.",
        points: [
            "File manager without upload limits, FTP sub-users",
            "Backup schedules and restores users run themselves",
            "Import full cPanel account backups",
        ],
        screenshot: "/img/openpanel-screenshots/files/backups_list-everything.png",
        alt: "Backups list",
        items: ["File Manager", "FTP", "Backups", "Disk & Inodes Usage", "Web Terminal", "cPanel Migration"],
    },
    {
        id: "users",
        title: "Users & resellers",
        lead: "Hosting plans, resellers and everything you do with customer accounts every day.",
        points: [
            "Plans with limits, upgrade or downgrade in one click",
            "Resellers with their own packages and users",
            "Log in as any user to see what they see",
        ],
        screenshot: "/img/openadmin-screenshots/plans/hosting_plans-list.png",
        alt: "Hosting plans in OpenAdmin",
        items: [
            "Suspend / Delete accounts",
            "Upgrade / Downgrade package",
            "Reseller accounts",
            "Dedicated IP address",
            "Impersonation",
            "Account settings",
            "Resource usage",
        ],
    },
    {
        id: "server",
        title: "Server management",
        lead: "OpenAdmin runs one server or a whole cluster, and OpenCLI automates everything from the terminal.",
        points: [
            "Manage many servers from one OpenAdmin",
            "Service status, notifications and webhooks",
            "Every action available as an opencli command",
        ],
        screenshot: "/img/openadmin-screenshots/multi-server/cluster-page.png",
        alt: "Multi-server cluster in OpenAdmin",
        items: [
            "Multi-Server",
            "Simple server deployment",
            "DNS Clustering",
            "Feature Manager",
            "Edit configuration",
            "Resource usage statistics",
            "Service management",
            "Smart Notifications",
            "Webhooks",
            "Terminal Commands",
            "Custom ports",
            "Custom domains",
        ],
    },
    {
        id: "interface",
        title: "Interface",
        lead: "A fast, simple panel your customers can use without a manual, with your brand on it.",
        points: [
            "Themes, your logo, colors and fonts",
            "Dark mode, search and keyboard shortcuts",
            "Works the same on phone, tablet and desktop",
        ],
        screenshot: "/img/openpanel-screenshots/themes/modern.png",
        alt: "OpenPanel with the Modern theme",
        dark: false,
        items: [
            "Branding",
            "Fully responsive",
            "Dark Mode",
            "Server Info",
            "Advanced Search",
            "Multi-language support",
            "Keyboard Shortcuts",
        ],
    },
    {
        id: "billing",
        title: "Billing & API",
        lead: "Sell hosting with the billing software you already use, or build on the API and MCP server.",
        points: [
            "Modules for WHMCS, Blesta, FOSSBilling, ClientExec, WISECP and ClientFence",
            "REST API for everything in OpenAdmin",
            "MCP server for AI agents",
        ],
        screenshot: "/img/openadmin-screenshots/settings/api_server-enabled.png",
        alt: "OpenAdmin API with the interactive reference",
        items: ["WHMCS", "Blesta", "FOSSBilling", "ClientExec", "WISECP", "ClientFence", "API", "MCP"],
    },
];

// a few entries point at old or missing pages
const URL_FIX: Record<string, string> = {
    "Resource usage management": "/docs/admin/plans/hosting_plans/",
    PostgreSQL: "/docs/panel/postgresql/databases/",
    "Terminal Commands": "/docs/opencli/intro/",
    "Edit configuration": "/docs/admin/settings/general/",
};

type Feature = {
    name: string;
    description: string;
    url?: string;
    icon?: React.FC<any>;
    enterprise: boolean;
    category: string;
};

const all = Object.values(data).flat() as any[];
const byName = new Map(all.map((i) => [i.name, i]));
// the billing modules share the WHMCS icon, a letter tells them apart
const sharedIcon = new Set(["Blesta", "FOSSBilling", "ClientExec", "WISECP", "ClientFence"]);

const FEATURES: Feature[] = CATEGORIES.flatMap((c) =>
    c.items
        .map((n) => byName.get(n))
        .filter(Boolean)
        .map((i) => ({
            name: i.name,
            description: i.description,
            url: URL_FIX[i.name] ?? (i.url || undefined),
            icon: sharedIcon.has(i.name) ? undefined : i.icon,
            enterprise: i.status === "enterprise",
            category: c.id,
        })),
);

const link = "text-refine-react-light-link dark:text-refine-react-dark-link hover:underline";

const Badge = ({ enterprise }: { enterprise: boolean }) => (
    <span
        className={clsx(
            "rounded-full px-2 py-0.5 text-[11px] font-semibold whitespace-nowrap",
            enterprise
                ? "bg-amber-100 text-amber-800 dark:bg-transparent dark:border dark:border-amber-400/60 dark:text-amber-300"
                : "bg-emerald-100 text-emerald-800 dark:bg-transparent dark:border dark:border-emerald-400/60 dark:text-emerald-300",
        )}
    >
        {enterprise ? "Enterprise" : "Community"}
    </span>
);

const FeatureCard = ({ f }: { f: Feature }) => {
    const Icon = f.icon;
    return (
        <div className="flex gap-3 rounded-xl border border-gray-200 dark:border-gray-700 p-4 bg-gray-0 dark:bg-gray-900">
            <div className="shrink-0 w-9 h-9 rounded-lg bg-gray-100 dark:bg-gray-800 flex items-center justify-center overflow-hidden">
                {Icon ? (
                    <Icon className="w-6 h-6" />
                ) : (
                    <span className="text-sm font-bold text-gray-700 dark:text-gray-200">{f.name[0]}</span>
                )}
            </div>
            <div className="flex flex-col gap-1 min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{f.name}</span>
                    <Badge enterprise={f.enterprise} />
                </div>
                <p className="text-sm text-gray-600 dark:text-gray-400 m-0 line-clamp-3">{f.description}</p>
                {f.url && (
                    <a href={f.url} className={clsx("text-xs", link)}>
                        Docs
                    </a>
                )}
            </div>
        </div>
    );
};

const Features: React.FC = () => {
    const [query, setQuery] = useState("");
    const [edition, setEdition] = useState<Edition>("all");

    const shown = useMemo(() => {
        const q = query.trim().toLowerCase();
        return FEATURES.filter(
            (f) =>
                (edition === "all" || (edition === "enterprise") === f.enterprise) &&
                (!q || f.name.toLowerCase().includes(q) || f.description.toLowerCase().includes(q)),
        );
    }, [query, edition]);

    const filtering = query.trim() !== "" || edition !== "all";

    // #section links open on that section, the header and filter bar are sticky so leave room for them
    useEffect(() => {
        const go = () => {
            const id = decodeURIComponent(window.location.hash.slice(1));
            const el = id && CATEGORIES.some((c) => c.id === id) ? document.getElementById(id) : null;
            if (el) window.scrollTo({ top: el.getBoundingClientRect().top + window.scrollY - 180 });
        };
        go();
        // screenshots load lazily and push sections down, so land again once they're in
        const t = setTimeout(go, 600);
        window.addEventListener("hashchange", go);
        return () => {
            clearTimeout(t);
            window.removeEventListener("hashchange", go);
        };
    }, []);

    return (
        <CommonLayout
            description="Everything OpenPanel does: per-account isolation, web servers, PHP, email, databases, backups, resellers, multi-server and billing integrations."
            image="/img/og/features.png"
        >
            <Head>
                <title>Features | OpenPanel</title>
                <meta property="og:title" content="Features | OpenPanel" />
                <html data-page="features" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="not-prose flex flex-col gap-10 pt-8 lg:pt-16 pb-32 max-w-[928px] landing-lg:max-w-[1232px] w-full mx-auto px-4">
                    <div className="flex flex-col gap-5">
                        <h1 className="text-3xl sm:text-5xl font-bold text-gray-900 dark:text-gray-0 m-0 max-w-[820px]">
                            A hosting panel where every account gets its own server stack
                        </h1>
                        <p className="text-base sm:text-lg text-gray-600 dark:text-gray-400 m-0 max-w-[720px]">
                            Web server, database, PHP and limits per account, isolated in containers. Everything your
                            customers expect from a control panel, without them stepping on each other.
                        </p>
                        <div className="flex flex-wrap gap-3">
                            <a href="/install/" className="rounded-full px-5 py-2.5 text-sm font-semibold bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900 hover:brightness-110 no-underline">
                                Install
                            </a>
                            <a href="/demo/" className="rounded-full px-5 py-2.5 text-sm font-semibold border border-gray-300 dark:border-gray-700 text-gray-900 dark:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800 no-underline">
                                Live demo
                            </a>
                            <a href="/enterprise/" className="rounded-full px-5 py-2.5 text-sm font-semibold border border-gray-300 dark:border-gray-700 text-gray-900 dark:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800 no-underline">
                                Compare editions
                            </a>
                        </div>
                        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 mt-2">
                            {[
                                ["Isolated", "containers for every account"],
                                ["~53 MB", "RAM for OpenPanel at idle"],
                                [`${FEATURES.length}+`, "features, each one on or off per user"],
                            ].map(([big, small]) => (
                                <div key={big} className="rounded-xl border border-gray-200 dark:border-gray-700 p-4">
                                    <div className="text-2xl font-bold text-gray-900 dark:text-gray-0">{big}</div>
                                    <div className="text-sm text-gray-600 dark:text-gray-400">{small}</div>
                                </div>
                            ))}
                        </div>
                    </div>

                    <div className="sticky top-[72px] z-[5] -mx-4 px-4 py-3 bg-gray-0/95 dark:bg-gray-900/95 backdrop-blur border-y border-gray-200 dark:border-gray-800 flex flex-col gap-3">
                        <div className="flex flex-col sm:flex-row gap-3 sm:items-center">
                            <input
                                type="search"
                                value={query}
                                onChange={(e) => setQuery(e.target.value)}
                                placeholder="Search features, e.g. backups"
                                aria-label="Search features"
                                className="w-full sm:max-w-[320px] rounded-lg border border-gray-300 dark:border-gray-700 bg-gray-0 dark:bg-gray-900 text-gray-900 dark:text-gray-100 placeholder:text-gray-400 px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-refine-blue/40"
                            />
                            <div className="inline-flex rounded-lg border border-gray-300 dark:border-gray-700 p-0.5 self-start">
                                {(["all", "community", "enterprise"] as Edition[]).map((e) => (
                                    <button
                                        key={e}
                                        type="button"
                                        onClick={() => setEdition(e)}
                                        className={clsx(
                                            "px-3 py-1.5 text-xs font-semibold rounded-md capitalize",
                                            edition === e
                                                ? "bg-gray-900 text-gray-0 dark:bg-gray-0 dark:text-gray-900"
                                                : "text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100",
                                        )}
                                    >
                                        {e}
                                    </button>
                                ))}
                            </div>
                        </div>
                        {!filtering && (
                            <nav className="flex gap-2 overflow-x-auto text-xs whitespace-nowrap -mb-1 pb-1">
                                {CATEGORIES.map((c) => (
                                    <a
                                        key={c.id}
                                        href={`#${c.id}`}
                                        className="rounded-full px-3 py-1 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 hover:border-gray-400 no-underline"
                                    >
                                        {c.title}
                                    </a>
                                ))}
                            </nav>
                        )}
                    </div>

                    {filtering ? (
                        <section className="flex flex-col gap-4">
                            <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                                {shown.length} {shown.length === 1 ? "feature" : "features"}
                            </p>
                            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                                {shown.map((f) => (
                                    <FeatureCard key={f.name} f={f} />
                                ))}
                            </div>
                        </section>
                    ) : (
                        CATEGORIES.map((c, idx) => (
                            <section key={c.id} id={c.id} className="flex flex-col gap-6 scroll-mt-[180px]">
                                <div
                                    className={clsx(
                                        "grid gap-6 items-center",
                                        c.screenshot && "lg:grid-cols-2",
                                    )}
                                >
                                    <div className={clsx("flex flex-col gap-3", c.screenshot && idx % 2 === 1 && "lg:order-2")}>
                                        <h2 className="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-gray-0 m-0">{c.title}</h2>
                                        <p className="text-base text-gray-600 dark:text-gray-400 m-0">{c.lead}</p>
                                        <ul className="m-0 p-0 list-none flex flex-col gap-2">
                                            {c.points.map((p) => (
                                                <li key={p} className="flex gap-2 text-sm text-gray-800 dark:text-gray-200">
                                                    <span className="text-refine-blue dark:text-refine-cyan-alt font-bold">✓</span>
                                                    {p}
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                    {c.screenshot && (
                                        <div className="rounded-xl border border-gray-200 dark:border-gray-700 overflow-hidden shadow-sm bg-gray-50 dark:bg-gray-800 aspect-[16/10]">
                                            <img
                                                src={c.screenshot}
                                                alt={c.alt}
                                                loading="lazy"
                                                className={clsx("w-full h-full object-cover object-left-top", c.dark !== false && "dark:hidden")}
                                            />
                                            {c.dark !== false && (
                                                <img
                                                    src={c.screenshot.replace(/\.png$/, "_dark.png")}
                                                    alt={c.alt}
                                                    loading="lazy"
                                                    className="w-full h-full object-cover object-left-top hidden dark:block"
                                                />
                                            )}
                                        </div>
                                    )}
                                </div>
                                <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                                    {FEATURES.filter((f) => f.category === c.id).map((f) => (
                                        <FeatureCard key={f.name} f={f} />
                                    ))}
                                </div>
                            </section>
                        ))
                    )}

                    <section className="rounded-2xl border border-gray-300 dark:border-gray-700 p-6 sm:p-8 flex flex-col sm:flex-row gap-4 sm:items-center sm:justify-between">
                        <div className="flex flex-col gap-1">
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Try it on your own server</h2>
                            <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                                One command installs everything. See what's coming next on the{" "}
                                <a href="/roadmap/" className={link}>roadmap</a>.
                            </p>
                        </div>
                        <div className="flex gap-3">
                            <a href="/install/" className="rounded-full px-5 py-2.5 text-sm font-semibold bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900 hover:brightness-110 no-underline">
                                Install
                            </a>
                            <a href="/demo/" className="rounded-full px-5 py-2.5 text-sm font-semibold border border-gray-300 dark:border-gray-700 text-gray-900 dark:text-gray-100 no-underline">
                                Live demo
                            </a>
                        </div>
                    </section>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Features;
