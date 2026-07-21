import React, { useState, ChangeEvent } from "react";
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
    openpanelui: {
        cpu: 0.5,
        ramGb: 0.5,
        diskGb: 2.6,
        label: "OpenPanel UI",
        description: "Web-based user interface for managing hosting accounts.",
    },
    openadminui: {
        cpu: 0,
        ramGb: 0.02,
        diskGb: 0.025,
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

const simpleIconProps = {
    width: 24,
    height: 24,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth: 1.5,
    strokeLinecap: "round" as const,
    strokeLinejoin: "round" as const,
};

const UsersIcon = () => (
    <svg {...simpleIconProps}>
        <circle cx="9" cy="8" r="3" />
        <path d="M3.5 20c0-3 2.5-5.5 5.5-5.5s5.5 2.5 5.5 5.5" />
        <circle cx="17" cy="8.5" r="2.3" />
        <path d="M15.7 14.7c2.5.3 4.3 2.5 4.3 5.3" />
    </svg>
);

const LayersIcon = () => (
    <svg {...simpleIconProps}>
        <path d="M12 3l9 4.5-9 4.5-9-4.5L12 3Z" />
        <path d="M3 12l9 4.5 9-4.5" />
        <path d="M3 16.5l9 4.5 9-4.5" />
    </svg>
);

const SparkleIcon = () => (
    <svg {...simpleIconProps}>
        <path d="M12 3l1.8 5.2L19 10l-5.2 1.8L12 17l-1.8-5.2L5 10l5.2-1.8L12 3Z" />
        <path d="M19 15l.7 2 2 .7-2 .7-.7 2-.7-2-2-.7 2-.7.7-2Z" />
    </svg>
);

const gradientAccentA = clsx(
    "text-transparent bg-clip-text bg-gradient-to-r text-gradient-to-r",
    "from-[#0FBDBD] to-[#26D97F]",
);

const gradientAccentB = clsx(
    "text-transparent bg-clip-text bg-gradient-to-r text-gradient-to-r",
    "from-[#FF9933] to-[#FF4C4D]",
);

const sectionHeading = clsx(
    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
    "tracking-tight",
    "p-0",
    "dark:text-gray-0 text-gray-900",
);

const sectionAccent = clsx(
    "font-semibold",
    "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
    "text-refine-blue drop-shadow-[0_0_30px_rgba(0,128,255,0.3)]",
);

const cardShell = clsx(
    "not-prose",
    "p-4 landing-sm:p-10",
    "flex",
    "flex-col landing-sm:flex-row",
    "items-start",
    "gap-6",
    "dark:bg-landing-noise",
    "dark:bg-gray-800 bg-gray-50",
    "rounded-2xl landing-sm:rounded-3xl",
);

const fieldLabel = clsx(
    "text-sm font-semibold",
    "dark:text-gray-0 text-gray-900",
    "mb-2 block",
);

const fieldHint = clsx(
    "text-sm",
    "mt-2",
    "dark:text-gray-400 text-gray-600",
);

const fieldInput = clsx(
    "w-full",
    "px-4 py-3",
    "text-sm",
    "dark:text-gray-0 text-gray-900",
    "dark:bg-gray-900 bg-white",
    "border dark:border-gray-700 border-gray-200",
    "rounded-xl",
    "outline-none",
    "focus:border-refine-blue dark:focus:border-refine-cyan-alt",
);

const checkboxInput = clsx(
    "w-4 h-4",
    "accent-refine-blue dark:accent-refine-cyan-alt",
);

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

    const renderAddonGrid = (addonDefs: Record<string, AddonCost>) =>
        Object.entries(addonDefs).map(([key, addon]) => (
            <label
                key={key}
                htmlFor={key}
                className={clsx(
                    "not-prose cursor-pointer",
                    "flex flex-col gap-3",
                    "p-6",
                    "rounded-2xl landing-sm:rounded-3xl",
                    "border",
                    "transition-colors duration-150",
                    addons[key]
                        ? clsx(
                              "dark:bg-refine-cyan-alt/10 bg-refine-blue/10",
                              "dark:border-refine-cyan-alt border-refine-blue",
                          )
                        : clsx(
                              "dark:bg-landing-noise dark:bg-gray-800 bg-gray-50",
                              "dark:border-gray-700 border-gray-200",
                          ),
                )}
            >
                <div className={clsx("flex items-center justify-between")}>
                    <span
                        className={clsx(
                            "flex items-center gap-2",
                            "text-lg font-semibold",
                            "dark:text-gray-0 text-gray-900",
                        )}
                    >
                        <SparkleIcon />
                        {addon.label}
                    </span>
                    <input
                        type="checkbox"
                        id={key}
                        name={key}
                        checked={addons[key]}
                        onChange={handleAddonChange}
                        className={checkboxInput}
                    />
                </div>
                <p className={clsx("m-0 text-sm", "dark:text-gray-400 text-gray-600")}>
                    {addon.description}
                </p>
                <p
                    className={clsx(
                        "m-0 text-xs font-semibold",
                        "dark:text-gray-500 text-gray-500",
                    )}
                >
                    +{addon.cpu} core &middot; +{addon.ramGb}GB RAM &middot; +
                    {addon.diskGb}GB disk
                </p>
            </label>
        ));

    const resultStats = [
        { value: `${recommendedDiskGb} GB`, label: "Minimum disk usage needed" },
        { value: `${recommendedCpu}`, label: "CPU cores" },
        { value: `${recommendedRamGb} GB`, label: "Memory" },
        {
            value: isFree ? "Free" : `€${monthlyPrice.toFixed(2)}`,
            label: isFree ? "OpenPanel license" : "OpenPanel license / month",
        },
    ];

    return (
        <CommonLayout description="Estimate the CPU, RAM, and disk you need to host a given number of users on OpenPanel, based on real benchmarks, plus your OpenPanel license cost.">
            <Head title="Resource Calculator | OpenPanel">
                <html data-page="calculator" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div
                    className={clsx(
                        "flex flex-col",
                        "gap-16 landing-sm:gap-20 landing-md:gap-28",
                        "w-full max-w-[592px] landing-sm:max-w-[656px] landing-md:max-w-[896px] landing-lg:max-w-[1200px]",
                        "px-2 landing-sm:px-0",
                        "pt-8 landing-sm:pt-12 landing-lg:pt-20",
                        "pb-16 landing-sm:pb-24 landing-md:pb-32",
                        "mx-auto",
                        "not-prose",
                    )}
                >
                    {/* Hero */}
                    <div className={clsx("px-4 landing-md:px-10")}>
                        <h1
                            className={clsx(
                                "text-3xl landing-sm:text-[40px] landing-sm:leading-[48px]",
                                "font-semibold",
                                "tracking-tight",
                                "text-center",
                                "max-w-3xl mx-auto",
                                "p-0",
                                "dark:text-gray-0 text-gray-900",
                            )}
                        >
                            Find the right{" "}
                            <span className={gradientAccentA}>server size</span> for
                            your{" "}
                            <span className={gradientAccentB}>hosting business</span>.
                        </h1>
                        <p
                            className={clsx(
                                "mt-4 landing-sm:mt-6",
                                "max-w-lg mx-auto",
                                "text-base text-center",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            Estimate the CPU, RAM, and disk you need based on real
                            OpenPanel benchmarks, and see what your OpenPanel license
                            would cost.
                        </p>
                    </div>

                    {/* Users */}
                    <div className={clsx("w-full")}>
                        <div className={clsx("px-4 landing-md:px-10")}>
                            <h2 className={sectionHeading}>
                                Tell us about your{" "}
                                <span className={sectionAccent}>users</span>.
                            </h2>
                            <p
                                className={clsx(
                                    "mt-4 landing-sm:mt-6",
                                    "max-w-md",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                How many hosting accounts, and how big each one's plan
                                is.
                            </p>
                        </div>

                        <div
                            className={clsx(
                                "mt-8 landing-sm:mt-12",
                                "px-4 landing-md:px-10",
                            )}
                        >
                            <div className={cardShell}>
                                <div className={clsx("hidden landing-sm:block")}>
                                    <UsersIcon />
                                </div>
                                <div
                                    className={clsx(
                                        "w-full",
                                        "grid",
                                        "grid-cols-1 landing-sm:grid-cols-3",
                                        "gap-6",
                                    )}
                                >
                                    <div>
                                        <label htmlFor="users" className={fieldLabel}>
                                            Number of users
                                        </label>
                                        <input
                                            id="users"
                                            type="number"
                                            min={1}
                                            value={users}
                                            onChange={(e) =>
                                                setUsers(
                                                    Math.max(1, Number(e.target.value)),
                                                )
                                            }
                                            className={fieldInput}
                                        />
                                        <div className={fieldHint}>
                                            Hosting accounts you plan to run.
                                        </div>
                                    </div>
                                    <div>
                                        <label htmlFor="cpu-per-user" className={fieldLabel}>
                                            CPU cores / user plan
                                        </label>
                                        <input
                                            id="cpu-per-user"
                                            type="number"
                                            min={0}
                                            max={16}
                                            step={0.5}
                                            value={cpuPerUser}
                                            onChange={(e) =>
                                                setCpuPerUser(
                                                    Math.max(
                                                        0,
                                                        Math.min(16, Number(e.target.value)),
                                                    ),
                                                )
                                            }
                                            className={fieldInput}
                                        />
                                        <div className={fieldHint}>
                                            Cap assigned per user, not typical usage.
                                        </div>
                                    </div>
                                    <div>
                                        <label htmlFor="ram-per-user" className={fieldLabel}>
                                            RAM (GB) / user plan
                                        </label>
                                        <input
                                            id="ram-per-user"
                                            type="number"
                                            min={0}
                                            max={16}
                                            step={0.5}
                                            value={ramPerUser}
                                            onChange={(e) =>
                                                setRamPerUser(
                                                    Math.max(
                                                        0,
                                                        Math.min(16, Number(e.target.value)),
                                                    ),
                                                )
                                            }
                                            className={fieldInput}
                                        />
                                        <div className={fieldHint}>
                                            Cap assigned per user, not typical usage.
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Services */}
                    <div className={clsx("w-full")}>
                        <div className={clsx("px-4 landing-md:px-10")}>
                            <h2 className={sectionHeading}>
                                Choose your{" "}
                                <span className={sectionAccent}>services</span>.
                            </h2>
                            <p
                                className={clsx(
                                    "mt-4 landing-sm:mt-6",
                                    "max-w-md",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                Shared, read-only images used by your users - each one
                                is downloaded once and reused, no matter how many users
                                run it.
                            </p>
                        </div>

                        <div
                            className={clsx(
                                "mt-8 landing-sm:mt-12",
                                "px-4 landing-md:px-10",
                            )}
                        >
                            <div className={cardShell}>
                                <div className={clsx("hidden landing-sm:block")}>
                                    <LayersIcon />
                                </div>
                                <div
                                    className={clsx(
                                        "w-full",
                                        "grid",
                                        "grid-cols-2 landing-sm:grid-cols-3",
                                        "gap-6",
                                    )}
                                >
                                    <div>
                                        <label htmlFor="database" className={fieldLabel}>
                                            Database
                                        </label>
                                        <select
                                            id="database"
                                            value={database}
                                            onChange={(e) => setDatabase(e.target.value)}
                                            className={fieldInput}
                                        >
                                            <option value="mariadb">MariaDB</option>
                                            <option value="mysql">MySQL</option>
                                        </select>
                                    </div>
                                    <div>
                                        <label htmlFor="webserver" className={fieldLabel}>
                                            Web server
                                        </label>
                                        <select
                                            id="webserver"
                                            value={webserver}
                                            onChange={(e) => setWebserver(e.target.value)}
                                            className={fieldInput}
                                        >
                                            <option value="nginx">Nginx</option>
                                            <option value="apache">Apache</option>
                                            <option value="openlitespeed">
                                                OpenLiteSpeed
                                            </option>
                                            <option value="openresty">OpenResty</option>
                                        </select>
                                    </div>
                                    <div>
                                        <label htmlFor="php-versions" className={fieldLabel}>
                                            PHP versions
                                        </label>
                                        <input
                                            id="php-versions"
                                            type="number"
                                            min={0}
                                            max={15}
                                            value={phpVersions}
                                            onChange={(e) =>
                                                setPhpVersions(
                                                    Math.max(
                                                        0,
                                                        Math.min(15, Number(e.target.value)),
                                                    ),
                                                )
                                            }
                                            className={fieldInput}
                                        />
                                        <div className={fieldHint}>
                                            How many distinct PHP versions your users
                                            need, each its own shared image.
                                        </div>
                                    </div>
                                    <div>
                                        <label htmlFor="caching" className={fieldLabel}>
                                            Caching
                                        </label>
                                        <select
                                            id="caching"
                                            value={caching}
                                            onChange={(e) => setCaching(e.target.value)}
                                            className={fieldInput}
                                        >
                                            <option value="none">None</option>
                                            <option value="memcached">Memcached</option>
                                            <option value="valkey">Valkey</option>
                                            <option value="redis">Redis</option>
                                        </select>
                                    </div>
                                    <div
                                        className={clsx(
                                            "flex items-center gap-2",
                                            "landing-sm:mt-7",
                                        )}
                                    >
                                        <input
                                            type="checkbox"
                                            id="cron"
                                            checked={cronEnabled}
                                            onChange={(e) =>
                                                setCronEnabled(e.target.checked)
                                            }
                                            className={checkboxInput}
                                        />
                                        <label htmlFor="cron" className="font-semibold">
                                            Cron
                                        </label>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Core Community services */}
                    <div className={clsx("w-full")}>
                        <div className={clsx("px-4 landing-md:px-10")}>
                            <h2 className={sectionHeading}>
                                Core{" "}
                                <span className={sectionAccent}>Community</span>{" "}
                                services.
                            </h2>
                            <p
                                className={clsx(
                                    "mt-4 landing-sm:mt-6",
                                    "max-w-md",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                Included with every OpenPanel install, free to use, and
                                each one can be disabled if you don't need it.
                            </p>
                        </div>

                        <div
                            className={clsx(
                                "mt-8 landing-sm:mt-12",
                                "px-4 landing-md:px-10",
                                "grid",
                                "grid-cols-1 landing-sm:grid-cols-3",
                                "gap-4 landing-sm:gap-6",
                            )}
                        >
                            {renderAddonGrid(COMMUNITY_ADDONS)}
                        </div>
                    </div>

                    {/* Enterprise services */}
                    <div className={clsx("w-full")}>
                        <div className={clsx("px-4 landing-md:px-10")}>
                            <h2 className={sectionHeading}>
                                Optional{" "}
                                <span className={sectionAccent}>Enterprise</span>{" "}
                                services.
                            </h2>
                            <p
                                className={clsx(
                                    "mt-4 landing-sm:mt-6",
                                    "max-w-md",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                Shared, system-wide services available on the
                                Enterprise Edition.
                            </p>
                        </div>

                        <div
                            className={clsx(
                                "mt-8 landing-sm:mt-12",
                                "px-4 landing-md:px-10",
                                "grid",
                                "grid-cols-1 landing-sm:grid-cols-3",
                                "gap-4 landing-sm:gap-6",
                            )}
                        >
                            {renderAddonGrid(ENTERPRISE_ADDONS)}
                        </div>
                    </div>

                    {/* Recommended VPS */}
                    <div className={clsx("w-full")}>
                        <div className={clsx("px-4 landing-md:px-10")}>
                            <h2 className={sectionHeading}>
                                Your{" "}
                                <span className={sectionAccent}>recommended</span> VPS.
                            </h2>
                        </div>

                        <div
                            className={clsx(
                                "mt-8 landing-sm:mt-12",
                                "px-4 landing-md:px-10",
                            )}
                        >
                            <div
                                className={clsx(
                                    "grid",
                                    "grid-cols-2 landing-sm:grid-cols-4",
                                    "gap-4",
                                )}
                            >
                                {resultStats.map((stat, index) => (
                                    <div
                                        key={index}
                                        className={clsx(
                                            "not-prose",
                                            "flex flex-col items-center text-center",
                                            "gap-2",
                                            "p-6",
                                            "dark:bg-landing-noise",
                                            "dark:bg-gray-800 bg-gray-50",
                                            "rounded-2xl",
                                        )}
                                    >
                                        <div
                                            className={clsx(
                                                "text-2xl landing-sm:text-3xl",
                                                "font-semibold",
                                                "dark:text-refine-cyan-alt text-refine-blue",
                                            )}
                                        >
                                            {stat.value}
                                        </div>
                                        <div
                                            className={clsx(
                                                "text-sm",
                                                "dark:text-gray-400 text-gray-600",
                                            )}
                                        >
                                            {stat.label}
                                        </div>
                                    </div>
                                ))}
                            </div>
                            <p
                                className={clsx(
                                    "mt-6",
                                    "text-base",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                This assumes typical usage, where most users are idle
                                most of the time - the same pattern we see across real
                                OpenPanel servers. If every single user maxed out their
                                plan at the exact same moment, you'd need{" "}
                                <strong className="dark:text-gray-0 text-gray-900">
                                    {round(maxCpu, 1)}
                                </strong>{" "}
                                cores and{" "}
                                <strong className="dark:text-gray-0 text-gray-900">
                                    {round(maxRamGb, 0.5)}
                                </strong>{" "}
                                GB RAM - oversell between the two is normal and
                                expected, the same way every shared hosting platform is
                                priced.
                            </p>
                            <p
                                className={clsx(
                                    "text-sm",
                                    "dark:text-gray-500 text-gray-500",
                                )}
                            >
                                Figures are based on internal OpenPanel benchmarks
                                (idle overhead of ~{PER_USER_IDLE_CPU} core and ~15MB
                                RAM per user, ~100MB disk per user, and a ~
                                {SYSTEM_BASE_DISK_GB}GB shared system image baseline)
                                and are an estimate - actual usage varies with real
                                traffic.{" "}
                                <Link to="/calculator/benchmark/">
                                    See how we benchmarked these numbers
                                </Link>
                                .
                            </p>
                        </div>
                    </div>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Calculator;
