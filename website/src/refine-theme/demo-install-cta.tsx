import React from "react";
import clsx from "clsx";
import Link from "@docusaurus/Link";
import { ENTERPRISE_TRIAL_URL } from "./enterprise-get-in-touch-button";

const INSTALL_COMMAND = "bash <(curl -sSL https://openpanel.org)";

const track = (event: string) => {
    if (typeof window !== "undefined" && typeof (window as any).gtag !== "undefined") {
        (window as any).gtag("event", event);
    }
};

const link = "text-refine-react-light-link dark:text-refine-react-dark-link hover:underline";

export const DemoInstallCta = ({ className }: { className?: string }) => {
    const [copied, setCopied] = React.useState(false);
    const timeoutRef = React.useRef<ReturnType<typeof setTimeout> | null>(null);

    const onCopy = () => {
        navigator.clipboard?.writeText(INSTALL_COMMAND);
        track("demo_install_copy");
        setCopied(true);
        if (timeoutRef.current) clearTimeout(timeoutRef.current);
        timeoutRef.current = setTimeout(() => setCopied(false), 2000);
    };

    return (
        <section
            className={clsx(
                "flex flex-col gap-4",
                "border-t border-gray-200 dark:border-gray-700 pt-10",
                className,
            )}
        >
            <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-0 m-0">Install it on your own server</h2>

            <p className="m-0 text-base text-gray-600 dark:text-gray-400">
                On a fresh Ubuntu 24.04 server, log in as root and run the command below. It takes about five minutes,
                asks no questions and prints the OpenAdmin login at the end. Community edition is free; for Enterprise
                features you can start a{" "}
                <Link to={ENTERPRISE_TRIAL_URL} onClick={() => track("trial_click_cta_demo")} className={link}>
                    30-day trial
                </Link>{" "}
                without a credit card.
            </p>

            <div
                className={clsx(
                    "flex items-center gap-3",
                    "rounded-lg border border-gray-300 dark:border-gray-700",
                    "bg-gray-50 dark:bg-gray-900",
                    "pl-4 pr-1.5 py-1.5",
                    "font-jetBrains-mono text-sm",
                )}
            >
                <code
                    className={clsx(
                        "flex-1 overflow-x-auto whitespace-nowrap",
                        "!bg-transparent !border-0 !p-0 text-gray-900 dark:text-gray-100",
                    )}
                >
                    {INSTALL_COMMAND}
                </code>
                <button
                    type="button"
                    onClick={onCopy}
                    className={clsx(
                        "shrink-0 appearance-none cursor-pointer",
                        "rounded-md px-3 py-1.5",
                        "text-sm font-sans",
                        "border border-gray-300 dark:border-gray-600",
                        "bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100",
                        "hover:bg-gray-100 dark:hover:bg-gray-700",
                    )}
                >
                    {copied ? "Copied" : "Copy"}
                </button>
            </div>

            <p className="m-0 text-sm text-gray-600 dark:text-gray-400">
                Need a domain, custom ports or no firewall? Use the{" "}
                <Link to="/install" className={link}>
                    install command generator
                </Link>
                .
            </p>
        </section>
    );
};
