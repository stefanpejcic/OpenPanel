import React from "react";

// No standardized MCP (Model Context Protocol) mark exists yet, so this
// uses a generic "plug/connector" concept instead of borrowing another
// company's logo to imply an endorsement that doesn't exist.
export const LandingHeroMcpIcon = (props: React.SVGProps<SVGSVGElement>) => (
    <svg
        width={24}
        height={64}
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
    >
        <path d="M9.785 6l8.215 8.215l-2.054 2.054a5.81 5.81 0 1 1 -8.215 -8.215l2.054 -2.054" />
        <path d="M4 20l3.5 -3.5" />
        <path d="M15 4l-3.5 3.5" />
        <path d="M20 9l-3.5 3.5" />
    </svg>
);
