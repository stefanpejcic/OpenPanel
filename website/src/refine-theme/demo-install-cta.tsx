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

const highlights = [
    { value: "5 min", label: "from fresh server to live panel" },
    { value: "1", label: "command, no prompts" },
    { value: "$0", label: "Community edition, forever" },
    { value: "30 days", label: "Enterprise trial, no card" },
];

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
                "relative overflow-hidden",
                "rounded-2xl landing-md:rounded-3xl",
                "p-6 landing-sm:p-10 landing-md:p-14",
                "bg-gray-50 landing-md:bg-landing-wizard-option-bg-light landing-md:bg-landing-wizard-option-left landing-md:bg-landing-wizard-option",
                "text-left",
                className,
            )}
        >
            <div
                aria-hidden
                className="pointer-events-none absolute inset-0 hidden dark:block bg-gradient-to-br from-[#0b2a4a] via-[#0d1b2e] to-[#0a3b3b]"
            />
            <div
                aria-hidden
                className="pointer-events-none absolute -top-32 -right-32 h-80 w-80 rounded-full bg-[#26D97F] opacity-0 dark:opacity-20 blur-3xl"
            />
            <div
                aria-hidden
                className="pointer-events-none absolute -bottom-32 -left-32 h-80 w-80 rounded-full bg-[#0FBDBD] opacity-0 dark:opacity-20 blur-3xl"
            />

            <div className="relative flex flex-col gap-6">
                <span
                    className={clsx(
                        "self-start",
                        "rounded-full px-3 py-1",
                        "text-xs font-semibold uppercase tracking-wider",
                        "bg-refine-blue/10 text-refine-blue dark:bg-[#26D97F]/15 dark:text-[#26D97F]",
                    )}
                >
                    Liked what you saw?
                </span>

                <h2
                    className={clsx(
                        "m-0 font-semibold text-gray-900 dark:text-white",
                        "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px] landing-md:text-[40px] landing-md:leading-[48px]",
                    )}
                >
                    Now run it on your own server.{" "}
                    <span className="text-transparent bg-clip-text bg-gradient-to-r from-[#0FBDBD] to-[#26D97F]">
                        Live in 5 minutes.
                    </span>
                </h2>

                <p className="m-0 text-base landing-sm:text-lg text-gray-600 dark:text-gray-300 max-w-[640px]">
                    Grab a fresh Ubuntu 24.04 VPS, connect as root and paste this one command.
                    The installer needs no input and prints your OpenAdmin login when it finishes.
                </p>

                <div
                    className={clsx(
                        "flex items-center gap-3",
                        "rounded-xl",
                        "bg-white ring-1 ring-gray-200 dark:bg-black/50 dark:ring-white/10",
                        "pl-4 landing-sm:pl-6 pr-2 py-2",
                        "font-jetBrains-mono text-sm landing-sm:text-base",
                    )}
                >
                    <span className="select-none text-refine-blue dark:text-[#26D97F]">$</span>
                    <code
                        className={clsx(
                            "flex-1 overflow-x-auto whitespace-nowrap",
                            "!bg-transparent !border-0 !p-0 text-gray-900 dark:text-white",
                        )}
                    >
                        {INSTALL_COMMAND}
                    </code>
                    <button
                        type="button"
                        onClick={onCopy}
                        className={clsx(
                            "shrink-0 appearance-none border-none cursor-pointer",
                            "rounded-lg px-4 py-2",
                            "text-sm font-semibold font-sans",
                            "transition-colors duration-150",
                            copied
                                ? "bg-[#26D97F] text-gray-900"
                                : "bg-refine-blue text-white hover:brightness-110 dark:bg-white/10 dark:hover:bg-white/20",
                        )}
                    >
                        {copied ? "Copied!" : "Copy"}
                    </button>
                </div>

                <div className="flex flex-col landing-sm:flex-row landing-sm:items-center gap-4 landing-sm:gap-6 pt-2">
                    <Link
                        to={ENTERPRISE_TRIAL_URL}
                        onClick={() => track("trial_click_cta_demo")}
                        className={clsx(
                            "self-start",
                            "rounded-3xl",
                            "!text-gray-0 dark:!text-gray-900",
                            "bg-refine-blue dark:bg-gradient-to-r dark:from-[#0FBDBD] dark:to-[#26D97F]",
                            "transition-[filter] duration-150 ease-in-out",
                            "hover:brightness-110 hover:!no-underline",
                            "py-3 px-8",
                            "text-base font-semibold whitespace-nowrap",
                        )}
                    >
                        Start 30-day Enterprise trial
                    </Link>
                    <span className="text-sm text-gray-500 dark:text-gray-400">
                        No credit card required.{" "}
                        <Link
                            to="/install"
                            className="!text-refine-blue dark:!text-gray-200 underline dark:hover:!text-white"
                        >
                            Customize install options →
                        </Link>
                    </span>
                </div>

                <div
                    className={clsx(
                        "grid grid-cols-2 landing-md:grid-cols-4",
                        "gap-4 landing-md:gap-6",
                        "mt-4 pt-6 border-t border-gray-200 dark:border-white/10",
                    )}
                >
                    {highlights.map((h) => (
                        <div key={h.label} className="flex flex-col">
                            <span className="text-xl landing-sm:text-2xl font-semibold text-gray-900 dark:text-white">
                                {h.value}
                            </span>
                            <span className="text-sm text-gray-500 dark:text-gray-400">{h.label}</span>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
};
