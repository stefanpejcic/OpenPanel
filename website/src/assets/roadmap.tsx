import React from "react";
import {
    MonitorIcon,
    SelfHostedIcon,
    IdentityIcon,
    ComponentsIcon,
    WizardsIcon,
    ProvidersIcon,
} from "@site/src/components/landing/icons";

// what is being worked on right now, empty hides the section; update it once things ship
export const inProgress: { icon: React.ReactNode; title: string; description: string }[] = [
];

// Everything else that's planned. Keep this list up to date manually.
export const plannedFeatures = [
    {
        icon: <MonitorIcon />,
        title: "Export OpenAdmin settings",
    },
    {
        icon: <SelfHostedIcon />,
        title: "Ongoing improvements in container isolation",
    },
    {
        icon: <IdentityIcon />,
        title: "Reseller rebranding",
    },
    {
        icon: <ComponentsIcon />,
        title: "Terraform provider",
    },
    {
        icon: <WizardsIcon />,
        title: "Ansible collection",
    },
    {
        icon: <ProvidersIcon />,
        title: "AI assistant: OpenAI Responses API (/v1/responses)",
    },
];
