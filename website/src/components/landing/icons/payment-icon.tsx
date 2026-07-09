import clsx from "clsx";
import * as React from "react";
import { SVGProps } from "react";

export const PaymentIcon = (props: SVGProps<SVGSVGElement>) => (
    <svg
        xmlns="http://www.w3.org/2000/svg"
        width={64}
        height={64}
        viewBox="0 0 64 64"
        fill="none"
        {...props}
        className={clsx(
            props.className,
            "dark:text-refine-yellow text-refine-orange",
        )}
    >
        <rect
            width={64}
            height={64}
            fill="url(#payment-a)"
            fillOpacity={0.4}
            rx={16}
        />
        <rect
            width={63}
            height={63}
            x={0.5}
            y={0.5}
            stroke="url(#payment-b)"
            strokeOpacity={0.5}
            rx={15.5}
        />
        <path
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M18 25a3 3 0 0 1 3-3h22a3 3 0 0 1 3 3v14a3 3 0 0 1-3 3H21a3 3 0 0 1-3-3V25ZM18 29h28M23 37h6"
        />
        <defs>
            <radialGradient
                id="payment-a"
                cx={0}
                cy={0}
                r={1}
                gradientTransform="rotate(45) scale(90.5097)"
                gradientUnits="userSpaceOnUse"
            >
                <stop stopColor="currentColor" />
                <stop offset={1} stopColor="currentColor" stopOpacity={0.25} />
            </radialGradient>
            <radialGradient
                id="payment-b"
                cx={0}
                cy={0}
                r={1}
                gradientTransform="rotate(45) scale(90.5097)"
                gradientUnits="userSpaceOnUse"
            >
                <stop stopColor="currentColor" />
                <stop offset={0.5} stopColor="currentColor" stopOpacity={0} />
                <stop offset={1} stopColor="currentColor" stopOpacity={0.25} />
            </radialGradient>
        </defs>
    </svg>
);
