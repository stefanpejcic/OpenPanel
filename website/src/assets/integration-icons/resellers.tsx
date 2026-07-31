import * as React from "react";
import { SVGProps } from "react";

const SvgResellers = (props: SVGProps<SVGSVGElement>) => (
    <svg
        width={24}
        height={24}
        viewBox="0 0 24 24"
        stroke="currentColor"
        fill="none"
        stroke-linecap="round"
        stroke-linejoin="round"
        xmlns="http://www.w3.org/2000/svg"
        {...props}
    >
  <path d="M10 13a2 2 0 1 0 4 0a2 2 0 0 0 -4 0" />
  <path d="M8 21v-1a2 2 0 0 1 2 -2h4a2 2 0 0 1 2 2v1" />
  <path d="M15 5a2 2 0 1 0 4 0a2 2 0 0 0 -4 0" />
  <path d="M17 10h2a2 2 0 0 1 2 2v1" />
  <path d="M5 5a2 2 0 1 0 4 0a2 2 0 0 0 -4 0" />
  <path d="M3 13v-1a2 2 0 0 1 2 -2h2" />
</svg>
);

export default SvgResellers;
