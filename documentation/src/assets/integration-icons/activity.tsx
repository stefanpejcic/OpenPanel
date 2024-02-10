import * as React from "react";
import { SVGProps } from "react";

const SvgActivity = (props: SVGProps<SVGSVGElement>) => (
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
  <path d="M3 12h4l3 8l4 -16l3 8h4" />
</svg>
);

export default SvgActivity;