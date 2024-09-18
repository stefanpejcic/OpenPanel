import * as React from "react";
import { SVGProps } from "react";

const SvgResponsive = (props: SVGProps<SVGSVGElement>) => (
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
  <path stroke="none" d="M0 0h24v24H0z" fill="none" />
  <path d="M6 5a2 2 0 0 1 2 -2h8a2 2 0 0 1 2 2v14a2 2 0 0 1 -2 2h-8a2 2 0 0 1 -2 -2v-14z" />
  <path d="M11 4h2" />
  <path d="M12 17v.01" />
</svg>
);

export default SvgResponsive;