import {
    DocumentsIcon,
    IntegrationsIcon,
    TutorialIcon,
    ExamplesIcon,
    AwesomeIcon,
    DiscordIcon,
    ContributeIcon,
    RefineWeekIcon,
    HackathonsIcon,
    AboutUsIcon,
    StoreIcon,
    MeetIcon,
    BlogIcon,
    NewBadgeIcon,
} from "../icons/popover";

export type NavbarPopoverItemType = {
    isPopover: true;
    label: string;
    items: {
        label: string;
        description: string;
        link: string;
        icon: React.FC;
    }[];
};

export type NavbarItemType = {
    isPopover?: false;
    label: string;
    icon?: React.FC;
    href?: string;
};

export type MenuItemType = NavbarPopoverItemType | NavbarItemType;


export const MENU_ITEMS: MenuItemType[] = [
  {
    isPopover: true,
    label: "Resources",
    items: [
      {
        label: "Documentation",
        description: "Everything you need to get started.",
        link: "/docs/",
        icon: DocumentsIcon,
      },
      {
        label: "Live Demo",
        description: "Give OpenPanel or OpenAdmin a try.",
        link: "/demo",
        icon: TutorialIcon,
      },
      {
        label: "Install Command",
        description: "Generate command to customize installation.",
        link: "/install",
        icon: AwesomeIcon,
      },
      {
        label: "Integrations",
        description: "WHMCS, FOSSBilling, Blesta.",
        link: "/features",
        icon: IntegrationsIcon,
      },
      {
        label: "How-to Guides",
        description: "Most common tasks on OpenPanel.",
        link: "/docs/articles/intro/",
        icon: ExamplesIcon,
      },
      {
        label: "Changelog",
        description: "Latest version and changes.",
        link: "/docs/changelog/intro/",
        icon: BlogIcon,
      },
    ],
  },
  {
    isPopover: false,
    label: "Live Demo",
    href: "/demo",
    icon: NewBadgeIcon,
  },
  {
    isPopover: false,
    label: "Enterprise",
    href: "/beta",
  },
  {
    isPopover: true,
    label: "Support",
    items: [
      {
        label: "Forums",
        description: "Join our Community forums!",
        link: "https://community.openpanel.org/",
        icon: DiscordIcon,
      },
      {
        label: "Discord",
        description: "Join us on Discord.",
        link: "https://discord.openpanel.com",
        icon: StoreIcon,
      },
      {
        label: "Contacy Us",
        description: "Email us for any questions.",
        link: "mailto:info@openpanel.com",
        icon: MeetIcon,
      },
    ],
  },
];
