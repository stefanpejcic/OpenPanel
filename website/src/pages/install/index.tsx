import React, { useMemo, useState } from "react";
import clsx from "clsx";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";

type Kind = "text" | "number" | "password" | "toggle";

interface InstallOption {
    flag: string;
    label: string;
    kind: Kind;
    description: string;
    placeholder?: string;
    docs?: string;
    // the value goes in single quotes so spaces and special characters reach the installer as typed
    quote?: boolean;
    validate?: (v: string) => string | undefined;
}

interface Group {
    title: string;
    description: string;
    options: InstallOption[];
}

const domainRe = /^(?!-)(?:[A-Za-z0-9-]{1,63}\.)+[A-Za-z]{2,}$/;
const port = (v: string) =>
    /^\d+$/.test(v) && +v >= 1000 && +v <= 30000 ? undefined : "A port between 1000 and 30000.";

const GROUPS: Group[] = [
    {
        title: "License & cluster",
        description: "Activate Enterprise right away, or install this server straight into a multi-server cluster.",
        options: [
            {
                flag: "key",
                label: "Enterprise license key",
                kind: "text",
                placeholder: "enterprise-...",
                description: "Activates the Enterprise edition during the install.",
                docs: "/docs/admin/license/#license-key",
                validate: (v) => (v.startsWith("enterprise-") ? undefined : "The key starts with enterprise-."),
            },
            {
                flag: "join-cluster",
                label: "Join a cluster",
                kind: "text",
                placeholder: "Code from the master server",
                description: "Paste the code from Generate install command on the master server. Needs a license key.",
                docs: "/docs/admin/system/multi-server/add-servers/#install-a-new-server-straight-into-the-cluster",
            },
        ],
    },
    {
        title: "Domains & ports",
        description: "By default both panels open on the server IP, OpenAdmin on port 2087 and OpenPanel on 2083.",
        options: [
            {
                flag: "domain",
                label: "Domain for both panels",
                kind: "text",
                placeholder: "server.example.com",
                description: "Domain used to open OpenAdmin and OpenPanel, with an SSL certificate.",
                docs: "/docs/admin/settings/general/#domain",
                validate: (v) => (domainRe.test(v) ? undefined : "Enter a domain like server.example.com."),
            },
            {
                flag: "panel-domain",
                label: "Separate OpenPanel domain",
                kind: "text",
                placeholder: "panel.example.com",
                description: "Use a different domain for OpenPanel only.",
                docs: "/docs/admin/settings/general/#domain",
                validate: (v) => (domainRe.test(v) ? undefined : "Enter a domain like panel.example.com."),
            },
            {
                flag: "admin-port",
                label: "OpenAdmin port",
                kind: "number",
                placeholder: "2087",
                description: "Custom port for OpenAdmin.",
                docs: "/docs/admin/settings/general/#ports",
                validate: port,
            },
            {
                flag: "user-port",
                label: "OpenPanel port",
                kind: "number",
                placeholder: "2083",
                description: "Custom port for OpenPanel.",
                docs: "/docs/admin/settings/general/#ports",
                validate: port,
            },
        ],
    },
    {
        title: "Admin account",
        description: "Left empty, a random username and password are generated and shown when the install finishes.",
        options: [
            {
                flag: "username",
                label: "Admin username",
                kind: "text",
                placeholder: "Random",
                description: "Username for the first OpenAdmin administrator.",
                docs: "/docs/admin/accounts/administrators/",
                quote: true,
                validate: (v) => (/^[a-zA-Z0-9_]{5,30}$/.test(v) ? undefined : "5 to 30 letters, numbers or underscores."),
            },
            {
                flag: "password",
                label: "Admin password",
                kind: "password",
                placeholder: "Random",
                description: "Password for that administrator.",
                docs: "/docs/admin/accounts/administrators/",
                quote: true,
                validate: (v) => (v.includes("'") ? "Single quotes can't be used here." : undefined),
            },
        ],
    },
    {
        title: "Security",
        description: "CSF firewall and the CorazaWAF are installed by default, ImunifyAV is optional.",
        options: [
            {
                flag: "skip-firewall",
                label: "Skip the firewall",
                kind: "toggle",
                description: "Don't install Sentinel Firewall (CSF). Make sure the server has another firewall.",
                docs: "/docs/admin/security/firewall/",
            },
            {
                flag: "no-waf",
                label: "Skip the WAF",
                kind: "toggle",
                description: "Don't install CorazaWAF and keep it off for new domains.",
                docs: "/docs/admin/security/waf/",
            },
            {
                flag: "imunifyav",
                label: "Install ImunifyAV",
                kind: "toggle",
                description: "Install and set up the ImunifyAV malware scanner.",
                docs: "/docs/admin/security/imunify/",
            },
        ],
    },
    {
        title: "Advanced",
        description: "For special setups and troubleshooting, the defaults are right for a normal install.",
        options: [
            {
                flag: "swap",
                label: "Swap size (GB)",
                kind: "number",
                placeholder: "Sized by RAM",
                description: "Swap file size, 1 to 10 GB. By default it's sized by RAM: twice the RAM up to 2 GB, then 4 GB up to 8 GB of RAM and 8 GB up to 32 GB. Servers with more than 32 GB of RAM get no swap unless you set a size.",
                validate: (v) => (/^\d+$/.test(v) && +v >= 1 && +v <= 10 ? undefined : "A whole number from 1 to 10."),
            },
            {
                flag: "post-install",
                label: "Post-install script",
                kind: "text",
                placeholder: "/root/after.sh or https://...",
                description: "Path or URL of a script to run when the install finishes.",
                quote: true,
            },
            {
                flag: "skip-dns-server",
                label: "Skip the DNS server",
                kind: "toggle",
                description: "Don't set up the local BIND9 DNS server, for example when DNS is hosted elsewhere.",
                docs: "/docs/admin/domains/dns/",
            },
            {
                flag: "skip-requirements",
                label: "Skip the requirements check",
                kind: "toggle",
                description: "Install even if the server doesn't meet the minimum requirements.",
                docs: "/docs/articles/install-update/system-requirements/",
            },
            {
                flag: "skip-panel-check",
                label: "Skip the existing panel check",
                kind: "toggle",
                description: "Don't check whether another control panel is already installed.",
            },
            {
                flag: "skip-apt-update",
                label: "Skip the package update",
                kind: "toggle",
                description: "Don't update the system packages before installing.",
            },
            {
                flag: "selfsigned",
                label: "Self-signed certificate",
                kind: "toggle",
                description: "Generate and use a self-signed SSL certificate.",
            },
            {
                flag: "repair",
                label: "Repair",
                kind: "toggle",
                description: "Retry the install and overwrite everything from a previous attempt.",
            },
        ],
    },
];

const ALL = GROUPS.flatMap((g) => g.options);
const BASE = "bash <(curl -sSL https://openpanel.org)";

const fieldClass = clsx(
    "w-full",
    "rounded-lg",
    "border",
    "bg-gray-0 dark:bg-gray-900",
    "text-gray-900 dark:text-gray-100",
    "placeholder:text-gray-400 dark:placeholder:text-gray-500",
    "px-3 py-2",
    "text-sm",
    "outline-none focus:ring-2 focus:ring-refine-blue/40 dark:focus:ring-refine-cyan-alt/40",
);

const Install: React.FC = () => {
    const [values, setValues] = useState<Record<string, string | boolean>>({});
    const [copied, setCopied] = useState(false);

    const errors = useMemo(() => {
        const out: Record<string, string> = {};
        for (const o of ALL) {
            const v = values[o.flag];
            if (typeof v === "string" && v.trim() && o.validate) {
                const e = o.validate(v.trim());
                if (e) out[o.flag] = e;
            }
        }
        return out;
    }, [values]);

    const command = useMemo(() => {
        let cmd = BASE;
        for (const o of ALL) {
            const v = values[o.flag];
            if (o.kind === "toggle") {
                if (v === true) cmd += ` --${o.flag}`;
            } else if (typeof v === "string" && v.trim() && !errors[o.flag]) {
                cmd += o.quote ? ` --${o.flag}='${v.trim()}'` : ` --${o.flag}=${v.trim()}`;
            }
        }
        return cmd;
    }, [values, errors]);

    const changed = Object.values(values).some((v) => v === true || (typeof v === "string" && v.trim() !== ""));

    const copy = async () => {
        try {
            await navigator.clipboard.writeText(command);
            setCopied(true);
            setTimeout(() => setCopied(false), 1500);
        } catch {
            setCopied(false);
        }
    };

    const set = (flag: string, v: string | boolean) => setValues((s) => ({ ...s, [flag]: v }));

    return (
        <CommonLayout
            description="Generate a custom OpenPanel install command with your choice of domain, ports, firewall, DNS, and security options."
            image="/img/og/install.png"
        >
            <Head>
                <title>Install command generator | OpenPanel</title>
                <meta property="og:title" content="Install command generator | OpenPanel" />
                <meta property="og:image:width" content="1200" />
                <meta property="og:image:height" content="630" />
                <meta property="og:image:type" content="image/png" />
                <html data-page="install" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="not-prose flex-1 flex flex-col gap-8 pt-8 lg:pt-16 pb-32 max-w-[960px] w-full mx-auto px-4">
                    <div className="flex flex-col gap-3">
                        <h1 className="text-3xl sm:text-4xl font-bold text-gray-900 dark:text-gray-0 m-0">
                            Install OpenPanel
                        </h1>
                        <p className="text-base text-gray-600 dark:text-gray-400 m-0 max-w-[720px]">
                            Connect to a fresh server as root over SSH and run the command below. The defaults work for most
                            servers, change the options further down and the command updates as you go.
                        </p>
                    </div>

                    <div
                        className={clsx(
                            "sticky top-[72px] z-[5]",
                            "rounded-2xl",
                            "border border-gray-300 dark:border-gray-700",
                            "bg-gray-0/95 dark:bg-gray-900/95 backdrop-blur",
                            "shadow-sm",
                            "p-4 flex flex-col gap-3",
                        )}
                    >
                        <div className="flex items-center justify-between gap-3">
                            <span className="text-sm font-semibold text-gray-700 dark:text-gray-300">Install command</span>
                            <div className="flex items-center gap-2">
                                {changed && (
                                    <button
                                        type="button"
                                        onClick={() => setValues({})}
                                        className="rounded-full px-3 py-1.5 text-xs font-semibold border border-gray-300 dark:border-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
                                    >
                                        Reset
                                    </button>
                                )}
                                <button
                                    type="button"
                                    onClick={copy}
                                    className="rounded-full px-4 py-1.5 text-xs font-semibold bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900 hover:brightness-110"
                                >
                                    {copied ? "Copied" : "Copy"}
                                </button>
                            </div>
                        </div>
                        <code
                            data-testid="install-command"
                            className="block rounded-lg bg-gray-100 dark:bg-gray-800 px-4 py-3 font-mono text-sm text-gray-900 dark:text-gray-100 break-all whitespace-pre-wrap"
                        >
                            {command}
                        </code>
                    </div>

                    <div className="flex flex-col gap-6">
                        {GROUPS.map((g) => (
                            <section
                                key={g.title}
                                className="rounded-2xl border border-gray-300 dark:border-gray-700 p-5 sm:p-6 flex flex-col gap-5"
                            >
                                <div className="flex flex-col gap-1">
                                    <h2 className="text-lg font-semibold text-gray-900 dark:text-gray-0 m-0">{g.title}</h2>
                                    <p className="text-sm text-gray-600 dark:text-gray-400 m-0">{g.description}</p>
                                </div>
                                <div className="grid gap-5 sm:grid-cols-2">
                                    {g.options.map((o) => {
                                        const v = values[o.flag];
                                        const err = errors[o.flag];
                                        const meta = (
                                            <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
                                                <code className="rounded bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 font-mono text-gray-700 dark:text-gray-300">
                                                    --{o.flag}
                                                </code>
                                                {o.docs && (
                                                    <a
                                                        href={o.docs}
                                                        className="text-refine-react-light-link dark:text-refine-react-dark-link hover:underline"
                                                    >
                                                        Docs
                                                    </a>
                                                )}
                                            </div>
                                        );
                                        if (o.kind === "toggle") {
                                            return (
                                                <label
                                                    key={o.flag}
                                                    htmlFor={o.flag}
                                                    className={clsx(
                                                        "flex items-start gap-3 rounded-xl p-3 cursor-pointer",
                                                        "border",
                                                        v === true
                                                            ? "border-refine-blue dark:border-refine-cyan-alt bg-refine-blue-2-light dark:bg-refine-blue-2 dark:bg-opacity-10"
                                                            : "border-gray-200 dark:border-gray-700 hover:border-gray-400 dark:hover:border-gray-500",
                                                    )}
                                                >
                                                    <input
                                                        id={o.flag}
                                                        type="checkbox"
                                                        checked={v === true}
                                                        onChange={(e) => set(o.flag, e.target.checked)}
                                                        className="mt-1 h-4 w-4 shrink-0 accent-refine-blue"
                                                    />
                                                    <span className="flex flex-col gap-1">
                                                        <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">
                                                            {o.label}
                                                        </span>
                                                        <span className="text-sm text-gray-600 dark:text-gray-400">{o.description}</span>
                                                        {meta}
                                                    </span>
                                                </label>
                                            );
                                        }
                                        return (
                                            <div key={o.flag} className="flex flex-col gap-1.5">
                                                <label htmlFor={o.flag} className="text-sm font-semibold text-gray-900 dark:text-gray-100">
                                                    {o.label}
                                                </label>
                                                <input
                                                    id={o.flag}
                                                    name={o.flag}
                                                    type={o.kind}
                                                    value={typeof v === "string" ? v : ""}
                                                    placeholder={o.placeholder}
                                                    onChange={(e) => set(o.flag, e.target.value)}
                                                    aria-invalid={!!err}
                                                    autoComplete="off"
                                                    className={clsx(
                                                        fieldClass,
                                                        err
                                                            ? "border-red-500 dark:border-red-500"
                                                            : "border-gray-300 dark:border-gray-700",
                                                    )}
                                                />
                                                {err ? (
                                                    <span className="text-xs text-red-600 dark:text-red-400">{err}</span>
                                                ) : (
                                                    <span className="text-sm text-gray-600 dark:text-gray-400">{o.description}</span>
                                                )}
                                                {meta}
                                            </div>
                                        );
                                    })}
                                </div>
                            </section>
                        ))}
                    </div>

                    <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                        Not sure the server is big enough? See the{" "}
                        <a href="/docs/articles/install-update/system-requirements/" className="text-refine-react-light-link dark:text-refine-react-dark-link hover:underline">
                            system requirements
                        </a>{" "}
                        or the{" "}
                        <a href="/calculator/" className="text-refine-react-light-link dark:text-refine-react-dark-link hover:underline">
                            resource calculator
                        </a>
                        . After the install, follow the{" "}
                        <a href="/docs/admin/intro/#post-install-steps" className="text-refine-react-light-link dark:text-refine-react-dark-link hover:underline">
                            post-install steps
                        </a>
                        .
                    </p>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Install;
