import * as React from "react";
import { SVGProps } from "react";

const SvgSeparate = (props: SVGProps<SVGSVGElement>) => (
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
  <path d="M4 4h16" />
  <path d="M4 20h16" />
</svg>
);

export default SvgSeparate;