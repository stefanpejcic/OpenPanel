import * as React from "react";
import { SVGProps } from "react";

const SvgFeatureManager = (props: SVGProps<SVGSVGElement>) => (
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
  <path d="M14 12a2 2 0 1 0 4 0a2 2 0 1 0 -4 0" />
  <path d="M2 12a6 6 0 0 1 6 -6h8a6 6 0 0 1 6 6a6 6 0 0 1 -6 6h-8a6 6 0 0 1 -6 -6" />
</svg>
);

export default SvgFeatureManager;
