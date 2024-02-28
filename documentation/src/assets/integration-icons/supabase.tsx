import * as React from "react";
import { SVGProps } from "react";

const SvgSupabase = (props: SVGProps<SVGSVGElement>) => (
    <svg
        width={24}
        height={64}
        viewBox="0 0 512 512"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
    >
    <path stroke="#066da5" stroke-width="38" d="M296 226h42m-92 0h42m-91 0h42m-91 0h41m-91 0h42m8-46h41m8 0h42m7 0h42m-42-46h42" />
    <path fill="#066da5" d="m472 228s-18-17-55-11c-4-29-35-46-35-46s-29 35-8 74c-6 3-16 7-31 7H68c-5 19-5 145 133 145 99 0 173-46 208-130 52 4 63-39 63-39" />
    </svg>
        <defs>
            <linearGradient
                id={`${props?.id || ""}-supabase-a`}
                x1="23.1353"
                y1="23.4843"
                x2="40.3698"
                y2="30.7124"
                gradientUnits="userSpaceOnUse"
            >
                <stop stopColor="#249361" />
                <stop offset="1" stopColor="#3ECF8E" />
            </linearGradient>
            <linearGradient
                id={`${props?.id || ""}-supabase-b`}
                x1="15.4944"
                y1="13.0225"
                x2="23.3542"
                y2="27.8183"
                gradientUnits="userSpaceOnUse"
            >
                <stop />
                <stop offset="1" stopOpacity="0" />
            </linearGradient>
        </defs>
    </svg>
);

export default SvgSupabase;
