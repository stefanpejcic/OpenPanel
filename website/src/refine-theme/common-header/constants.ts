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
        label: "Products",
        items: [
            {
                label: "Community Edition",
                description: "Free web hosting panel for VPS and private use.",
                link: "/docs",
                icon: ExamplesIcon,
            },
            {
                label: "Enterprise Edition",
                description: "Premium server control panel for shared hosting.",
                link: "/blog",
                icon: IntegrationsIcon,
            },
        ],
    },
    {
       isPopover: false,
       label: "Enterprise",
       href: "/beta",
   icon: NewBadgeIcon,
    },
    {
        isPopover: true,
        label: "Community",
        items: [
            {
                label: "Documentation",
                description: "Everything you need to get started.",
                link: "/docs/",
                icon: DocumentsIcon,
            },
            {
                label: "Forums",
                description: "Join our growing community!",
                link: "https://community.openpanel.org/",
                icon: ContributeIcon,
            },
            {
                label: "Translations",
                description: "Help us improve OpenPanel!",
                link: "https://github.com/stefanpejcic/openpanel-translations",
                icon: HackathonsIcon,
            },
        ],
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
