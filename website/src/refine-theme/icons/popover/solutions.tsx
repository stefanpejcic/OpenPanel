import React from "react";

// Navbar "Solutions" popover icons, same 40px tinted-circle style as the
// other popover icons.

const Circle = ({ color }: { color: string }) => (
    <>
        <rect
            width="40"
            height="40"
            rx="20"
            fill="currentColor"
            className="text-white dark:text-gray-900"
        />
        <rect
            width="40"
            height="40"
            rx="20"
            fill="currentColor"
            className={color}
            fillOpacity="0.1"
        />
    </>
);

export const HostingProvidersIcon = (props: React.SVGProps<SVGSVGElement>) => (
    <svg
        width="40"
        height="40"
        viewBox="0 0 40 40"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
    >
        <Circle color="dark:text-refine-cyan-alt text-refine-cyan" />
        <g
            className="dark:text-refine-cyan-alt text-refine-cyan"
            stroke="currentColor"
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
        >
            <rect x="12.5" y="12.5" width="15" height="6" rx="1.5" />
            <rect x="12.5" y="21.5" width="15" height="6" rx="1.5" />
            <path d="M15.5 15.5h.01M15.5 24.5h.01M19 15.5h5.5M19 24.5h5.5" />
        </g>
    </svg>
);

export const AgenciesIcon = (props: React.SVGProps<SVGSVGElement>) => (
    <svg
        width="40"
        height="40"
        viewBox="0 0 40 40"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
    >
        <Circle color="dark:text-refine-orange text-refine-orange" />
        <g
            className="text-refine-orange"
            stroke="currentColor"
            strokeWidth="1.6"
            strokeLinecap="round"
            strokeLinejoin="round"
        >
            <rect x="12" y="16" width="16" height="11" rx="2" />
            <path d="M17 16v-1.5a1.5 1.5 0 0 1 1.5-1.5h3a1.5 1.5 0 0 1 1.5 1.5V16" />
            <path d="M12 21h16M20 20v2" />
        </g>
    </svg>
);

// Small factory for the remaining line icons in the same tinted-circle style.
const lineIcon = (tint: string, stroke: string, paths: React.ReactNode) => {
    const Icon = (props: React.SVGProps<SVGSVGElement>) => (
        <svg
            width="40"
            height="40"
            viewBox="0 0 40 40"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            {...props}
        >
            <Circle color={tint} />
            <g
                className={stroke}
                stroke="currentColor"
                strokeWidth="1.6"
                strokeLinecap="round"
                strokeLinejoin="round"
            >
                {paths}
            </g>
        </svg>
    );
    return Icon;
};

export const MobileAppIcon = lineIcon(
    "dark:text-refine-indigo-alt text-refine-indigo",
    "dark:text-refine-indigo-alt text-refine-indigo",
    <>
        <rect x="15" y="11.5" width="10" height="17" rx="2" />
        <path d="M19 25.5h2" />
    </>,
);

export const MigrationIcon = lineIcon(
    "dark:text-refine-green-alt text-refine-green",
    "dark:text-refine-green-alt text-refine-green",
    <>
        <path d="M12.5 17h13l-3.5-3.5M27.5 23h-13l3.5 3.5" />
    </>,
);

export const BillingIcon = lineIcon(
    "text-refine-yellow",
    "text-refine-yellow",
    <>
        <rect x="11.5" y="14" width="17" height="12" rx="2" />
        <path d="M11.5 18h17M15 22.5h3" />
    </>,
);

export const PartnerIcon = lineIcon(
    "text-refine-pink",
    "text-refine-pink",
    <>
        <circle cx="17" cy="17" r="3" />
        <circle cx="24.5" cy="18" r="2.25" />
        <path d="M11.5 27c.5-3 2.6-4.5 5.5-4.5s5 1.5 5.5 4.5M22.5 22.6c.6-.3 1.3-.4 2-.4 2.2 0 3.7 1.2 4 3.8" />
    </>,
);

export const CalculatorIcon = lineIcon(
    "dark:text-refine-blue-alt text-refine-blue",
    "dark:text-refine-blue-alt text-refine-blue",
    <>
        <rect x="13.5" y="11.5" width="13" height="17" rx="2" />
        <path d="M16.5 15h7M16.5 19.5h.01M20 19.5h.01M23.5 19.5h.01M16.5 23h.01M20 23h.01M23.5 23h.01" />
    </>,
);

export const CompareIcon = lineIcon(
    "dark:text-refine-cyan-alt text-refine-cyan",
    "dark:text-refine-cyan-alt text-refine-cyan",
    <>
        <path d="M20 12v16M14 15l-3 6h6l-3-6ZM26 15l-3 6h6l-3-6ZM14 15h12M17 28h6" />
    </>,
);
