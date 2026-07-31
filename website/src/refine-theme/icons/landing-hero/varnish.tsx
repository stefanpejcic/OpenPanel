import React from "react";

// Varnish's real mark is three overlapping circles of different sizes.
export const LandingHeroVarnishIcon = (
    props: React.SVGProps<SVGSVGElement>,
) => (
    <svg
        width={24}
        height={64}
        viewBox="0 0 24 24"
        fill="#0072BC"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
    >
        <circle cx="16.5" cy="7.5" r="6" />
        <circle cx="4" cy="10" r="2.5" />
        <circle cx="10.5" cy="19" r="4" />
    </svg>
);
