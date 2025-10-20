import {
    DocumentsIcon,
    IntegrationsIcon,
    TutorialIcon,
    ExamplesIcon,
    AwesomeIcon,
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
            label: "How-to Guides",
            description: "How to customize OpenPanel.",
            link: "/docs/",
            icon: TutorialIcon,
          },
          {
            label: "Features",
            description: "Discover OpenPanel features.",
            link: "/features/",
            icon: AwesomeIcon,
          },
          {
            label: "Changelog",
            description: "New features and bug fixes.",
            link: "/changelog/",
            icon: IntegrationsIcon,
          },
          {
            label: "Forums",
            description: "Join our growing community!",
            link: "https://community.openpanel.org/",
            icon: ContributeIcon,
          },
          {
            label: "Blog",
            description: "Articles about Linux and server stuff.",
            link: "/blog",
            icon: BlogIcon,
          },
        ],
      },
    {
       isPopover: false,
       label: "Features",
       href: "/features",
    },
    {
       isPopover: false,
       label: "Live Demo",
       href: "/demo",
       icon: NewBadgeIcon,
    },
  {
    isPopover: true,
    label: "Enterprise",
    items: [
      {
        label: "OpenPanel Enterprise",
        description: "Fixed pricing at 14.95€ monthly.",
        link: "/enterprise/",
        icon: StoreIcon,
      },
      {
        label: "Hosting Providers",
        description: "Hosting providers ❤️ OpenPanel.",
        link: "/hosting-providers",
        icon: ExamplesIcon,
      },
      {
        label: "Contact us",
        description: "Call us for any questions",
        link: "mailto:info@openpanel.com",
        icon: MeetIcon,
      },
  //  {
   //     isPopover: true,
  //      label: "Company",
  //      items: [
   //         {
  //              label: "About Us",
  //              description: "Team & company information.",
       //         link: "/about",
    //            icon: AboutUsIcon,
  //          },
     //       {
    //            label: "Become a Partner",
      //          description: "Help us spread the word!",
      //          link: "mailto:info@openpanel.co",
      //          icon: StoreIcon,
    //        },
     //       {
     //           label: "Meet OpenPanel",
     //           description: "Call us for any questions",
     //           link: "mailto:info@openpanel.co",
     //           icon: MeetIcon,
     //       },
    //    ],
 //   },
];
