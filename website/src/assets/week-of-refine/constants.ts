import { CardProps } from "@site/src/pages/week-of-refine";
import { StrapiWithText, SupabaseWithText } from "../integration-icons";

export const weekOfRefineCards: CardProps[] = [
    {
        title: "FOSSBilling",
        dateRange: "February 7, 2024",
        description:
            "FOSSBilling module for OpenPanel.",
        logo: StrapiWithText,
        bgLinearGradient:
            "bg-week-of-refine-strapi-card-light dark:bg-week-of-refine-strapi-card",
        link: "/fossbilling-module",
    },
    {
        title: "WHMCS",
        dateRange: "December 14, 2023",
        description:
            "WHMCS module for OpenPanel.",
        logo: SupabaseWithText,
        bgLinearGradient:
            "bg-week-of-refine-supabase-card-light dark:bg-week-of-refine-supabase-card",
        link: "/whmcs-module",
    },
];
