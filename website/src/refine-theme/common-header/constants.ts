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
       isPopover: false,
       label: "Install",
       href: "/install",
    },
    {
       isPopover: false,
       label: "Docs",
       href: "/docs/",
    },   
    {
       isPopover: false,
       label: "Forums",
       href: "https://community.openpanel.org/",
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
