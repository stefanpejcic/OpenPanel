import React from "react";
import Link from "@docusaurus/Link";
import clsx from "clsx";
import { translate } from "@docusaurus/Translate";

import type { NavbarPopoverItemType } from "./constants";

type MenuItemProps = {
  item: NavbarPopoverItemType["items"][0];
  variant?: "landing" | "blog";
};

export const MenuItem: React.FC<MenuItemProps> = ({
  item,
  variant = "landing",
}) => {
  const Icon = item.icon;

  return (
    <Link to={item.link} className="no-underline">
      <div
        className={clsx(
          "flex items-start",
          "p-4",
          "transition duration-150 ease-in-out",
          "rounded-lg",
          "hover:bg-gray-50",
          variant === "landing" && "dark:hover:bg-gray-800",
          variant === "blog" && "dark:hover:bg-gray-700",
        )}
      >
        <div className="shrink-0">
          <Icon />
        </div>
        <div className="ml-2">
          <div
            className={clsx(
              variant === "landing" && "text-gray-900 dark:text-white",
              variant === "blog" &&
              "text-refine-react-8 dark:text-refine-react-3",
              "font-semibold",
            )}
          >
            {translate({ message: item.label, id: `menu.label.${item.label}` })}
          </div>
          <div
            className={clsx(
              variant === "landing" && "text-gray-500 dark:text-gray-400",
              variant === "blog" &&
              "text-refine-react-5 dark:text-refine-react-4",
              "text-xs",
            )}
          >
            {item.description ? translate({ message: item.description, id: `menu.desc.${item.label}` }) : null}
          </div>
        </div>
      </div>
    </Link>
  );
};
