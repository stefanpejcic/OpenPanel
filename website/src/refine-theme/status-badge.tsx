import clsx from "clsx";
import React, { useEffect, useState } from "react";

const STATUS_PAGE = "https://status.openpanel.com/status/";
const STATUS_JSON = "https://status.openpanel.com/status/status.json";

type State = "loading" | "up" | "down" | "error";

const LABELS: Record<State, string> = {
    loading: "Checking status…",
    up: "All services are online",
    down: "Some services are down",
    error: "View service status",
};

/**
 * Native replacement for the status.openpanel.com iframe badge. The iframe's
 * Tailwind `dark:` styles follow the visitor's OS color scheme, not the
 * site's theme toggle, so on our dark footer the text rendered gray-on-dark
 * for anyone with a light-mode OS. Reads the same status.json (CORS *).
 * Kept in sync with the my.openpanel.com WHMCS footer (.op-status-badge).
 */
export const StatusBadge = ({ className }: { className?: string }) => {
    const [state, setState] = useState<State>("loading");

    useEffect(() => {
        let cancelled = false;
        const load = () =>
            fetch(STATUS_JSON, { cache: "no-store" })
                .then((r) => r.json())
                .then((d) => {
                    const ok = (d.domains || []).every(
                        (x: { http_ok: boolean; whois_ok: boolean }) =>
                            x.http_ok && x.whois_ok,
                    );
                    if (!cancelled) setState(ok ? "up" : "down");
                })
                .catch(() => {
                    if (!cancelled) setState("error");
                });
        load();
        const id = setInterval(load, 60000);
        return () => {
            cancelled = true;
            clearInterval(id);
        };
    }, []);

    return (
        <a
            href={STATUS_PAGE}
            target="_blank"
            rel="noopener noreferrer"
            className={clsx(
                "inline-flex items-center gap-2",
                "h-[30px] pl-2 pr-3",
                "rounded-md border",
                "text-sm font-medium whitespace-nowrap",
                "no-underline hover:no-underline",
                "transition-colors",
                "border-gray-200 bg-gray-50 text-gray-900 hover:bg-gray-100 hover:text-gray-900",
                "dark:border-gray-700 dark:bg-white/5 dark:text-gray-0 dark:hover:bg-white/10 dark:hover:text-gray-0",
                className,
            )}
        >
            <span
                aria-hidden="true"
                className={clsx(
                    "w-2.5 h-2.5 rounded-full",
                    state === "up" && "bg-[#22c55e] shadow-[0_0_0_3px_rgba(34,197,94,0.25)]",
                    (state === "down" || state === "error") &&
                        "bg-[#ef4444] shadow-[0_0_0_3px_rgba(239,68,68,0.25)]",
                    state === "loading" && "bg-gray-400 shadow-[0_0_0_3px_rgba(138,147,168,0.25)]",
                )}
            />
            <span>{LABELS[state]}</span>
        </a>
    );
};
