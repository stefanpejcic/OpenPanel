import * as React from "react";
import { SVGProps } from "react";

const SvgMigration = (props: SVGProps<SVGSVGElement>) => (
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
  <path d="M20 10h-16l5.5 -6" />
  <path d="M4 14h16l-5.5 6" />
</svg>
);

export default SvgMigration;
