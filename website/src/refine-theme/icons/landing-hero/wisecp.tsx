import React from "react";

// Simplified monogram inspired by the WISECP mark: a "W" on a blue
// rounded-square badge.
export const LandingHeroWisecpIcon = (
    props: React.SVGProps<SVGSVGElement>,
) => (
    <svg
        width={24}
        height={64}
        viewBox="0 0 28 28"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
    >
        <rect x={2} y={2} width={24} height={24} rx={7} fill="#1479D6" />
        <text
            x="50%"
            y="53%"
            textAnchor="middle"
            dominantBaseline="middle"
            fontFamily="system-ui, sans-serif"
            fontWeight={800}
            fontSize={12}
            fill="#fff"
        >
            W
        </text>
    </svg>
);
