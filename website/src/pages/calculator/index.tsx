import React, { useState, useEffect, useRef, ChangeEvent } from "react";
import Head from "@docusaurus/Head";
import Link from "@docusaurus/Link";
import clsx from "clsx";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";

// Fixed system baseline: mysql, redis, caddy running idle. OpenPanel UI is a
// separate toggle below since it can be disabled. Based on real telemetry
// (all 5 system containers combined used ~0.87GB RAM in production testing).
const SYSTEM_BASE_CPU = 0.5;
const SYSTEM_BASE_RAM_GB = 1;
// Disk footprint of UI/OpenAdmin/Caddy/phpMyAdmin/DNS now lives on their own
// addon entries below (so disabling one also drops its disk cost); this is
// what's left over for the base OS/system install itself.
const SYSTEM_BASE_DISK_GB = 0.9;

// Per-user overhead when idle (no active traffic), based on internal benchmarks.
// CPU figure is derived from real system-wide utilization (vmstat), not summed
// per-container stats, since container-count-driven kernel/network-namespace
// overhead doesn't show up in any single container's own CPU accounting.
const PER_USER_IDLE_RAM_MB = 15;
const PER_USER_IDLE_CPU = 0.25;
const PER_USER_DISK_MB = 100;

// Shared, read-only service images: downloaded once and reused by every user, regardless of user count.
const SERVICE_IMAGE_MB: Record<string, number> = {
    mariadb: 347,
    mysql: 962,
    nginx: 64,
    apache: 69,
    openlitespeed: 682,
    openresty: 158,
    memcached: 14,
    valkey: 44,
    redis: 98,
    cron: 25,
};

// Each selected PHP version is its own shared image, downloaded once regardless of user count.
const PHP_VERSION_MB = 330;

const ENTERPRISE_MONTHLY_PRICE = 14.95;

interface AddonCost {
    cpu: number;
    ramGb: number;
    diskGb: number;
    label: string;
    description: string;
}

const COMMUNITY_ADDONS: Record<string, AddonCost> = {
    // 2.0.28: openpanel/openpanel image 278 MB, whole container ~119 MB RAM, 0% CPU idle and ~15% of a core under load
    openpanelui: {
        cpu: 0.25,
        ramGb: 0.15,
        diskGb: 0.3,
        label: "OpenPanel UI",
        description: "Web-based user interface for managing hosting accounts.",
    },
    // 2.0.28: ~33 MB RAM, 34.5 MB on disk, plus the 11.4 MB opencli binary
    openadminui: {
        cpu: 0,
        ramGb: 0.035,
        diskGb: 0.05,
        label: "OpenAdmin UI",
        description: "Interface for admin users and resellers.",
    },
    caddy: {
        cpu: 0.5,
        ramGb: 0.5,
        diskGb: 0.125,
        label: "Caddy",
        description:
            "Web server and reverse proxy for hosted domains. Only disable for a DNS-only or Email-only server.",
    },
    dns: {
        cpu: 0.25,
        ramGb: 0.125,
        diskGb: 0.2,
        label: "DNS",
        description: "Local authoritative BIND9 DNS server for hosted domains.",
    },
    phpmyadmin: {
        cpu: 0.5,
        ramGb: 0.5,
        diskGb: 0.75,
        label: "phpMyAdmin",
        description: "Web-based MySQL/MariaDB administration.",
    },
};

const ENTERPRISE_ADDONS: Record<string, AddonCost> = {
    email: {
        cpu: 0.5,
        ramGb: 1,
        diskGb: 1,
        label: "Email",
        description: "Roundcube webmail + full mailserver for all hosted domains.",
    },
    ftp: {
        cpu: 0.1,
        ramGb: 0.25,
        diskGb: 0.1,
        label: "FTP",
        description: "FTP/SFTP access for hosted accounts.",
    },
    clamav: {
        cpu: 4,
        ramGb: 4,
        diskGb: 0.8,
        label: "ClamAV",
        description: "Antivirus scanning for uploaded files and email.",
    },
    imunifyav: {
        cpu: 1,
        ramGb: 2,
        diskGb: 0.6,
        label: "ImunifyAV",
        description: "Malware scanning and remediation for hosted websites.",
    },
};

const ALL_ADDONS: Record<string, AddonCost> = {
    ...COMMUNITY_ADDONS,
    ...ENTERPRISE_ADDONS,
};

const round = (value: number, step: number) => Math.ceil(value / step) * step;


const card = "rounded-2xl border border-gray-300 dark:border-gray-700 p-5 sm:p-6 flex flex-col gap-5";
const cardTitle = "text-lg font-semibold text-gray-900 dark:text-gray-0 m-0";
const cardLead = "text-sm text-gray-600 dark:text-gray-400 m-0";
const fieldLabel = "text-sm font-semibold text-gray-900 dark:text-gray-100";
const fieldHint = "text-sm text-gray-600 dark:text-gray-400";
const fieldInput = clsx(
    "w-full rounded-lg border border-gray-300 dark:border-gray-700",
    "bg-gray-0 dark:bg-gray-900 text-gray-900 dark:text-gray-100",
    "px-3 py-2 text-sm outline-none",
    "focus:ring-2 focus:ring-refine-blue/40 dark:focus:ring-refine-cyan-alt/40",
);
const toggleCard = (on: boolean) =>
    clsx(
        "flex items-start gap-3 rounded-xl p-3 cursor-pointer border",
        on
            ? "border-refine-blue dark:border-refine-cyan-alt bg-refine-blue-2-light dark:bg-refine-blue-2 dark:bg-opacity-10"
            : "border-gray-200 dark:border-gray-700 hover:border-gray-400 dark:hover:border-gray-500",
    );
const link = "text-refine-react-light-link dark:text-refine-react-dark-link hover:underline";

const Calculator: React.FC = () => {
    const [users, setUsers] = useState(10);
    const [cpuPerUser, setCpuPerUser] = useState(2);
    const [ramPerUser, setRamPerUser] = useState(2);

    const [database, setDatabase] = useState("mariadb");
    const [webserver, setWebserver] = useState("nginx");
    const [phpVersions, setPhpVersions] = useState(1);
    const [caching, setCaching] = useState("none");
    const [cronEnabled, setCronEnabled] = useState(true);

    const [addons, setAddons] = useState<Record<string, boolean>>({
        openpanelui: true,
        openadminui: true,
        caddy: true,
        dns: false,
        phpmyadmin: false,
        email: false,
        ftp: false,
        clamav: false,
        imunifyav: false,
    });

    const handleAddonChange = (e: ChangeEvent<HTMLInputElement>) => {
        const { name, checked } = e.target;
        setAddons((prev) => ({ ...prev, [name]: checked }));
    };

    const resultsRef = useRef<HTMLDivElement>(null);

    // Prefill from URL params (e.g. ?users=50&cpu=4&ram=4&database=mysql&email=1)
    // so a link can be shared with the calculator already set up for someone.
    useEffect(() => {
        const params = new URLSearchParams(window.location.search);
        if ([...params.keys()].length === 0) return;

        const getInt = (key: string, fallback: number, min: number, max: number) =>
            params.has(key)
                ? Math.max(min, Math.min(max, Number(params.get(key))))
                : fallback;

        const getBool = (key: string, fallback: boolean) =>
            params.has(key)
                ? params.get(key) === "1" || params.get(key) === "true"
                : fallback;

        const getEnum = <T extends string>(
            key: string,
            fallback: T,
            allowed: readonly T[],
        ) => {
            const value = params.get(key) as T | null;
            return value && allowed.includes(value) ? value : fallback;
        };

        setUsers((prev) => getInt("users", prev, 1, 100000));
        setCpuPerUser((prev) => getInt("cpu", prev, 0, 16));
        setRamPerUser((prev) => getInt("ram", prev, 0, 16));
        setDatabase((prev) =>
            getEnum("database", prev, ["mariadb", "mysql"] as const),
        );
        setWebserver((prev) =>
            getEnum(
                "webserver",
                prev,
                ["nginx", "apache", "openlitespeed", "openresty"] as const,
            ),
        );
        setPhpVersions((prev) => getInt("php", prev, 0, 15));
        setCaching((prev) =>
            getEnum(
                "caching",
                prev,
                ["none", "memcached", "valkey", "redis"] as const,
            ),
        );
        setCronEnabled((prev) => getBool("cron", prev));
        setAddons((prev) => {
            const next = { ...prev };
            Object.keys(ALL_ADDONS).forEach((key) => {
                next[key] = getBool(key, prev[key]);
            });
            return next;
        });

        resultsRef.current?.scrollIntoView({ behavior: "smooth", block: "start" });
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    // OpenLiteSpeed only supports one PHP version.
    useEffect(() => {
        if (webserver === "openlitespeed") {
            setPhpVersions(1);
        }
    }, [webserver]);

    // Keep the URL in sync with the current inputs, so the address bar is
    // always a shareable link to this exact configuration. Skipped on the
    // very first run so it doesn't clobber params a shared link came in with.
    const isFirstUrlSync = useRef(true);
    useEffect(() => {
        if (isFirstUrlSync.current) {
            isFirstUrlSync.current = false;
            return;
        }
        const params = new URLSearchParams();
        params.set("users", String(users));
        params.set("cpu", String(cpuPerUser));
        params.set("ram", String(ramPerUser));
        params.set("database", database);
        params.set("webserver", webserver);
        params.set("php", String(phpVersions));
        params.set("caching", caching);
        params.set("cron", cronEnabled ? "1" : "0");
        Object.keys(ALL_ADDONS).forEach((key) => {
            params.set(key, addons[key] ? "1" : "0");
        });
        window.history.replaceState(
            null,
            "",
            `${window.location.pathname}?${params.toString()}`,
        );
    }, [
        users,
        cpuPerUser,
        ramPerUser,
        database,
        webserver,
        phpVersions,
        caching,
        cronEnabled,
        addons,
    ]);

    const selectedServiceImages: string[] = [
        database,
        webserver,
        ...(caching !== "none" ? [caching] : []),
        ...(cronEnabled ? ["cron"] : []),
    ];

    const sharedImagesMb =
        selectedServiceImages.reduce(
            (sum, key) => sum + (SERVICE_IMAGE_MB[key] ?? 0),
            0,
        ) +
        phpVersions * PHP_VERSION_MB;

    const addonCpu = Object.entries(addons)
        .filter(([, enabled]) => enabled)
        .reduce((sum, [key]) => sum + ALL_ADDONS[key].cpu, 0);

    const addonRamGb = Object.entries(addons)
        .filter(([, enabled]) => enabled)
        .reduce((sum, [key]) => sum + ALL_ADDONS[key].ramGb, 0);

    const addonDiskGb = Object.entries(addons)
        .filter(([, enabled]) => enabled)
        .reduce((sum, [key]) => sum + ALL_ADDONS[key].diskGb, 0);

    // Each additional PHP version beyond the first adds a bit of CPU/RAM overhead.
    const extraPhpVersions = Math.max(0, phpVersions - 1);
    const extraPhpCpu = extraPhpVersions * 0.25;
    const extraPhpRamGb = extraPhpVersions * 0.25;

    // Caching adds a small CPU/RAM overhead when enabled.
    const cachingCpu = caching !== "none" ? 0.1 : 0;
    const cachingRamGb = caching !== "none" ? 0.1 : 0;

    const diskGb =
        SYSTEM_BASE_DISK_GB +
        addonDiskGb +
        sharedImagesMb / 1024 +
        (users * PER_USER_DISK_MB) / 1024;

    const typicalCpu =
        SYSTEM_BASE_CPU +
        addonCpu +
        extraPhpCpu +
        cachingCpu +
        users * PER_USER_IDLE_CPU;
    const typicalRamGb =
        SYSTEM_BASE_RAM_GB +
        addonRamGb +
        extraPhpRamGb +
        cachingRamGb +
        (users * PER_USER_IDLE_RAM_MB) / 1024;

    const maxCpu =
        SYSTEM_BASE_CPU + addonCpu + extraPhpCpu + cachingCpu + users * cpuPerUser;
    const maxRamGb =
        SYSTEM_BASE_RAM_GB +
        addonRamGb +
        extraPhpRamGb +
        cachingRamGb +
        users * ramPerUser;

    const recommendedCpu = round(typicalCpu, 1);
    const recommendedRamGb = round(typicalRamGb, 0.5);
    const recommendedDiskGb = round(diskGb, 5);

    const isFree =
        users <= 3 &&
        !addons.email &&
        !addons.dns &&
        database !== "mariadb" &&
        webserver !== "openlitespeed";
    const monthlyPrice = isFree ? 0 : ENTERPRISE_MONTHLY_PRICE;


    const [copied, setCopied] = useState(false);
    const copyLink = async () => {
        try {
            await navigator.clipboard.writeText(window.location.href);
            setCopied(true);
            setTimeout(() => setCopied(false), 1500);
        } catch {
            setCopied(false);
        }
    };

    const renderAddonGrid = (addonDefs: Record<string, AddonCost>) => (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {Object.entries(addonDefs).map(([key, addon]) => (
                <label key={key} htmlFor={key} className={toggleCard(addons[key])}>
                    <input
                        type="checkbox"
                        id={key}
                        name={key}
                        checked={addons[key]}
                        onChange={handleAddonChange}
                        className="mt-1 h-4 w-4 shrink-0 accent-refine-blue"
                    />
                    <span className="flex flex-col gap-1">
                        <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{addon.label}</span>
                        <span className="text-sm text-gray-600 dark:text-gray-400">{addon.description}</span>
                        <span className="text-xs text-gray-500 dark:text-gray-400 font-mono">
                            +{addon.cpu} core · +{addon.ramGb} GB RAM · +{addon.diskGb} GB disk
                        </span>
                    </span>
                </label>
            ))}
        </div>
    );

    const resultStats = [
        { value: `${recommendedCpu}`, label: "CPU cores" },
        { value: `${recommendedRamGb} GB`, label: "Memory" },
        { value: `${recommendedDiskGb} GB`, label: "Disk, minimum" },
        {
            value: isFree ? "Free" : `€${monthlyPrice.toFixed(2)}`,
            label: isFree ? "OpenPanel license" : "OpenPanel license / month",
        },
    ];

    const numberField = (
        id: string,
        label: string,
        hint: string,
        value: number,
        set: (n: number) => void,
        min: number,
        max: number,
        step = 1,
        disabled = false,
    ) => (
        <div className="flex flex-col gap-1.5">
            <label htmlFor={id} className={fieldLabel}>
                {label}
            </label>
            <input
                id={id}
                type="number"
                min={min}
                max={max}
                step={step}
                value={value}
                disabled={disabled}
                onChange={(e) => set(Math.max(min, Math.min(max, Number(e.target.value))))}
                className={clsx(fieldInput, disabled && "opacity-50 cursor-not-allowed")}
            />
            <span className={fieldHint}>{hint}</span>
        </div>
    );

    const selectField = (
        id: string,
        label: string,
        value: string,
        set: (v: string) => void,
        options: [string, string][],
    ) => (
        <div className="flex flex-col gap-1.5">
            <label htmlFor={id} className={fieldLabel}>
                {label}
            </label>
            <select id={id} value={value} onChange={(e) => set(e.target.value)} className={fieldInput}>
                {options.map(([v, l]) => (
                    <option key={v} value={v}>
                        {l}
                    </option>
                ))}
            </select>
        </div>
    );

    return (
        <CommonLayout
            description="Estimate the CPU, RAM, and disk you need to host a given number of users on OpenPanel, based on real benchmarks, plus your OpenPanel license cost."
            image="/img/og/calculator.png"
        >
            <Head>
                <title>Resource Calculator | OpenPanel</title>
                <meta property="og:title" content="Resource Calculator | OpenPanel" />
                <meta property="og:image:width" content="1200" />
                <meta property="og:image:height" content="630" />
                <meta property="og:image:type" content="image/png" />
                <html data-page="calculator" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="not-prose flex-1 flex flex-col gap-8 pt-8 lg:pt-16 pb-32 max-w-[928px] landing-lg:max-w-[1232px] w-full mx-auto px-4">
                    <div className="flex flex-col gap-3">
                        <h1 className="text-3xl sm:text-4xl font-bold text-gray-900 dark:text-gray-0 m-0">
                            Resource calculator
                        </h1>
                        <p className="text-base text-gray-600 dark:text-gray-400 m-0 max-w-[720px]">
                            Find the right server size for your hosting business. Estimates are based on real OpenPanel
                            benchmarks, change the options below and the numbers update as you go.
                        </p>
                    </div>

                    <div
                        ref={resultsRef}
                        className={clsx(
                            "sticky top-[72px] z-[5] scroll-mt-24",
                            "rounded-2xl border border-gray-300 dark:border-gray-700",
                            "bg-gray-0/95 dark:bg-gray-900/95 backdrop-blur shadow-sm",
                            "p-4 flex flex-col gap-3",
                        )}
                    >
                        <div className="flex items-center justify-between gap-3">
                            <span className="text-sm font-semibold text-gray-700 dark:text-gray-300">Recommended server</span>
                            <button
                                type="button"
                                onClick={copyLink}
                                className="rounded-full px-4 py-1.5 text-xs font-semibold bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900 hover:brightness-110"
                            >
                                {copied ? "Copied" : "Copy link"}
                            </button>
                        </div>
                        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                            {resultStats.map((s) => (
                                <div key={s.label} className="rounded-lg bg-gray-100 dark:bg-gray-800 px-4 py-3">
                                    <div className="text-xl sm:text-2xl font-bold text-gray-900 dark:text-gray-0" data-testid={`result-${s.label}`}>
                                        {s.value}
                                    </div>
                                    <div className="text-xs text-gray-600 dark:text-gray-400">{s.label}</div>
                                </div>
                            ))}
                        </div>
                    </div>

                    <section className={card}>
                        <div className="flex flex-col gap-1">
                            <h2 className={cardTitle}>Users</h2>
                            <p className={cardLead}>How many hosting accounts, and how big each one's plan is.</p>
                        </div>
                        <div className="grid gap-5 sm:grid-cols-3">
                            {numberField("users", "Number of users", "Hosting accounts you plan to run.", users, setUsers, 1, 100000)}
                            {numberField("cpu-per-user", "CPU cores / user plan", "Cap assigned per user, not typical usage.", cpuPerUser, setCpuPerUser, 0, 16, 0.5)}
                            {numberField("ram-per-user", "RAM (GB) / user plan", "Cap assigned per user, not typical usage.", ramPerUser, setRamPerUser, 0, 16, 0.5)}
                        </div>
                    </section>

                    <section className={card}>
                        <div className="flex flex-col gap-1">
                            <h2 className={cardTitle}>Services</h2>
                            <p className={cardLead}>
                                Shared, read-only images used by your users, each one is downloaded once and reused no matter how
                                many users run it.
                            </p>
                        </div>
                        <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
                            {selectField("database", "Database", database, setDatabase, [
                                ["mariadb", "MariaDB"],
                                ["mysql", "MySQL"],
                            ])}
                            {selectField("webserver", "Web server", webserver, setWebserver, [
                                ["nginx", "Nginx"],
                                ["apache", "Apache"],
                                ["openlitespeed", "OpenLiteSpeed"],
                                ["openresty", "OpenResty"],
                            ])}
                            {numberField(
                                "php-versions",
                                "PHP versions",
                                webserver === "openlitespeed"
                                    ? "OpenLiteSpeed only supports one PHP version."
                                    : "How many distinct PHP versions your users need, each its own shared image.",
                                phpVersions,
                                setPhpVersions,
                                0,
                                15,
                                1,
                                webserver === "openlitespeed",
                            )}
                            {selectField("caching", "Caching", caching, setCaching, [
                                ["none", "None"],
                                ["memcached", "Memcached"],
                                ["valkey", "Valkey"],
                                ["redis", "Redis"],
                            ])}
                            <label htmlFor="cron" className={clsx(toggleCard(cronEnabled), "self-start sm:mt-6")}>
                                <input
                                    type="checkbox"
                                    id="cron"
                                    checked={cronEnabled}
                                    onChange={(e) => setCronEnabled(e.target.checked)}
                                    className="mt-1 h-4 w-4 shrink-0 accent-refine-blue"
                                />
                                <span className="flex flex-col gap-1">
                                    <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">Cron</span>
                                    <span className="text-sm text-gray-600 dark:text-gray-400">Scheduled jobs for your users.</span>
                                </span>
                            </label>
                        </div>
                    </section>

                    <section className={card}>
                        <div className="flex flex-col gap-1">
                            <h2 className={cardTitle}>Core Community services</h2>
                            <p className={cardLead}>
                                Included with every OpenPanel install, free to use, and each one can be disabled if you don't need it.
                            </p>
                        </div>
                        {renderAddonGrid(COMMUNITY_ADDONS)}
                    </section>

                    <section className={card}>
                        <div className="flex flex-col gap-1">
                            <h2 className={cardTitle}>Optional Enterprise services</h2>
                            <p className={cardLead}>Shared, system-wide services available on the Enterprise Edition.</p>
                        </div>
                        {renderAddonGrid(ENTERPRISE_ADDONS)}
                    </section>

                    <section className="flex flex-col gap-3">
                        <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">How these numbers work</h2>
                        <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                            This assumes typical usage, where most users are idle most of the time, the same pattern we see across
                            real OpenPanel servers. If every single user maxed out their plan at the exact same moment, you'd need{" "}
                            <strong className="text-gray-900 dark:text-gray-0">{round(maxCpu, 1)}</strong> cores and{" "}
                            <strong className="text-gray-900 dark:text-gray-0">{round(maxRamGb, 0.5)}</strong> GB RAM. Oversell
                            between the two is normal and expected, the same way every shared hosting platform is priced.
                        </p>
                        <p className="text-sm text-gray-500 dark:text-gray-500 m-0">
                            Figures are based on internal OpenPanel benchmarks (idle overhead of ~{PER_USER_IDLE_CPU} core and
                            ~{PER_USER_IDLE_RAM_MB}MB RAM per user, ~{PER_USER_DISK_MB}MB disk per user, and a ~{SYSTEM_BASE_DISK_GB}GB
                            shared system image baseline) and are an estimate, actual usage varies with real traffic.{" "}
                            <Link to="/calculator/benchmark/" className={link}>
                                See how we benchmarked these numbers
                            </Link>
                            .
                        </p>
                    </section>

                    <section className="rounded-2xl border border-gray-300 dark:border-gray-700 p-6 sm:p-8 flex flex-col sm:flex-row gap-4 sm:items-center sm:justify-between">
                        <div className="flex flex-col gap-1">
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Got your server?</h2>
                            <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                                Check the{" "}
                                <a href="/docs/articles/install-update/system-requirements/" className={link}>
                                    system requirements
                                </a>{" "}
                                and install OpenPanel with one command.
                            </p>
                        </div>
                        <a
                            href="/install/"
                            className="self-start sm:self-auto rounded-full px-5 py-2.5 text-sm font-semibold bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900 hover:brightness-110 no-underline whitespace-nowrap"
                        >
                            Install command
                        </a>
                    </section>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Calculator;
