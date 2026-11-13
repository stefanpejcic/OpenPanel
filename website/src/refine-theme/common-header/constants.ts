import {
    DocumentsIcon,
    IntegrationsIcon,
    TutorialIcon,
    ExamplesIcon,
    AwesomeIcon,
    ContributeIcon,
    StoreIcon,
    BlogIcon,
    HostingProvidersIcon,
    AgenciesIcon,
    MobileAppIcon,
    MigrationIcon,
    BillingIcon,
    PartnerIcon,
    CalculatorIcon,
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
    // desktop popover layout
    columns?: 1 | 2;
    // which way the panel opens from its navbar label
    align?: "left" | "center" | "right";
    // compact row of text links under the cards (e.g. "Switch from")
    quickLinks?: {
        title: string;
        links: { label: string; to: string }[];
    };
    // bottom strip of the popover
    footer?: "github" | "community" | "contact" | "trial";
};

export type NavbarItemType = {
    isPopover?: false;
    label: string;
    icon?: React.FC;
    href?: string;
};

export type MenuItemType = NavbarPopoverItemType | NavbarItemType;

// Ordered along the buying journey: what it is, who it's for, what it
// costs, try it, then the reference material.
export const MENU_ITEMS: MenuItemType[] = [
  {
    isPopover: true,
    label: "Product",
    columns: 2,
    align: "left",
    footer: "trial",
    items: [
      {
        label: "Features",
        description: "See what's included, feature by feature.",
        link: "/features/",
        icon: AwesomeIcon,
      },
      {
        label: "Free Migration",
        description: "We move your accounts from cPanel, Plesk & more.",
        link: "/migrations/",
        icon: MigrationIcon,
      },
      {
        label: "Mobile App",
        description: "Manage your server from your phone.",
        link: "/mobile/",
        icon: MobileAppIcon,
      },
      {
        label: "Changelog",
        description: "Every release, documented as it ships.",
        link: "/docs/changelog/intro/",
        icon: IntegrationsIcon,
      },
    ],
  },
  {
    isPopover: true,
    label: "Solutions",
    columns: 2,
    align: "center",
    footer: "contact",
    items: [
      {
        label: "For Hosting Providers",
        description: "One licence, no add-ons, more margin.",
        link: "/solutions/for-hosting-providers/",
        icon: HostingProvidersIcon,
      },
      {
        label: "For Agencies",
        description: "Secure, isolated hosting for clients.",
        link: "/solutions/for-agencies/",
        icon: AgenciesIcon,
      },
      {
        label: "Billing Integrations",
        description: "WHMCS, Blesta, FOSSBilling, ClientExec, WiseCP.",
        link: "/billing/",
        icon: BillingIcon,
      },
    ],
    quickLinks: {
      title: "Switch from",
      links: [
        { label: "cPanel", to: "/cpanel-alternative/" },
        { label: "Plesk", to: "/plesk-alternative/" },
        { label: "DirectAdmin", to: "/directadmin-alternative/" },
        { label: "CyberPanel", to: "/cyberpanel-alternative/" },
        { label: "Enhance", to: "/enhance-alternative/" },
      ],
    },
  },
  {
    isPopover: true,
    label: "Pricing",
    columns: 2,
    align: "center",
    footer: "community",
    items: [
      {
        label: "OpenPanel Enterprise",
        description: "€14.95 per server / month. Unlimited accounts.",
        link: "/enterprise/",
        icon: StoreIcon,
      },
      {
        label: "OpenPanel Community",
        description: "Forever free, self-hosted.",
        link: "/community/",
        icon: ExamplesIcon,
      },
      {
        label: "NOC Partner Licences",
        description: "Volume pricing for hosting providers.",
        link: "/licenses-for-partners/",
        icon: PartnerIcon,
      },
      {
        label: "Resource Calculator",
        description: "Estimate hardware and licence cost.",
        link: "/calculator/",
        icon: CalculatorIcon,
      },
    ],
  },
  {
    isPopover: false,
    label: "Live Demo",
    href: "/demo",
  },
  {
    isPopover: true,
    label: "Resources",
    columns: 2,
    align: "right",
    footer: "github",
    items: [
      {
        label: "Documentation",
        description: "Setup, configuration, and API reference.",
        link: "/docs/",
        icon: DocumentsIcon,
      },
      {
        label: "How-to Guides",
        description: "Step-by-step guides for common tasks.",
        link: "/docs/articles/intro/",
        icon: TutorialIcon,
      },
      {
        label: "Discussions",
        description: "Share ideas, upvote features.",
        link: "https://github.com/stefanpejcic/OpenPanel/discussions",
        icon: ContributeIcon,
      },
      {
        label: "Blog",
        description: "Tips and guides on Linux hosting.",
        link: "/blog",
        icon: BlogIcon,
      },
    ],
  },
];
