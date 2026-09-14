import React from "react";

// Simplified monogram inspired by the ClientExec mark: a lowercase "e" in a
// blue circle wrapped by a dark globe-style ring open on the right.
export const LandingHeroClientExecIcon = (
    props: React.SVGProps<SVGSVGElement>,
) => (
    <svg
        width={24}
        height={64}
        viewBox="0 0 28 28"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
    >
        <circle
            cx={14}
            cy={14}
            r={11.5}
            fill="none"
            stroke="#26262B"
            strokeWidth={2.6}
            strokeLinecap="round"
            strokeDasharray="56.2 16.06"
        />
        <circle cx={14} cy={14} r={7.6} fill="#1479D6" />
        <text
            x="50%"
            y="53%"
            textAnchor="middle"
            dominantBaseline="middle"
            fontFamily="system-ui, sans-serif"
            fontWeight={800}
            fontSize={11}
            fill="#fff"
        >
            e
        </text>
    </svg>
);
