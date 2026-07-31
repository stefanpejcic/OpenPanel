import React from "react";

// FOSSBilling has no icon in dashboard-icons, simple-icons, or Iconify.
// Placeholder monogram badge instead; swap in a real asset if/when one is
// available.
export const LandingHeroFossBillingIcon = (
    props: React.SVGProps<SVGSVGElement>,
) => (
    <svg
        width={24}
        height={64}
        viewBox="0 0 24 24"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
    >
        <rect width={24} height={24} rx={6} fill="#22C55E" />
        <text
            x="50%"
            y="54%"
            textAnchor="middle"
            dominantBaseline="middle"
            fontFamily="system-ui, sans-serif"
            fontWeight={700}
            fontSize={11}
            fill="#fff"
        >
            FB
        </text>
    </svg>
);
