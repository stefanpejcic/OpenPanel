import * as React from "react";
import { SVGProps } from "react";

const SvgPorts = (props: SVGProps<SVGSVGElement>) => (
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
  <path d="M8 10v-7l-2 2" />
  <path d="M6 16a2 2 0 1 1 4 0c0 .591 -.601 1.46 -1 2l-3 3h4" />
  <path d="M15 14a2 2 0 1 0 2 -2a2 2 0 1 0 -2 -2" />
  <path d="M6.5 10h3" />
</svg>
);

export default SvgPorts;