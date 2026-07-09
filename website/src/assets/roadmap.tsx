import React from "react";
import {
    FolderIcon,
    WizardsIcon,
    SecurityIcon,
    NoVendorLockinIcon,
    PhpIcon,
    MonitorIcon,
    SelfHostedIcon,
    PaymentIcon,
} from "@site/src/components/landing/icons";

// The two items currently being worked on. Update this once they ship.
export const inProgress = [
    {
        icon: <FolderIcon />,
        title: "Backup restore GUI",
        description:
            "Browse existing backups and restore a website, database, or full account with a couple of clicks - no terminal required.",
    },
    {
        icon: <WizardsIcon />,
        title: "Autoinstallers",
        description:
            "One-click installers for popular apps - we're currently working on Drupal, Joomla, Moodle, and Nextcloud.",
    },
];

// Everything else that's planned. Keep this list up to date manually.
export const plannedFeatures = [
    {
        icon: <SecurityIcon />,
        title: "ImunifyAV page for end-users",
    },
    {
        icon: <NoVendorLockinIcon />,
        title: "Ubuntu 26 support",
    },
    {
        icon: <PhpIcon />,
        title: "PHP 8.6 version",
    },
    {
        icon: <MonitorIcon />,
        title: "Export OpenAdmin settings",
    },
    {
        icon: <SelfHostedIcon />,
        title: "Ongoing improvements in container isolation",
    },
    {
        icon: <PaymentIcon />,
        title: "Paymenter.org module",
    },
];
