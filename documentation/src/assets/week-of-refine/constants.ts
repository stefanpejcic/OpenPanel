import { CardProps } from "@site/src/pages/week-of-refine";
import { StrapiWithText, SupabaseWithText } from "../integration-icons";

export const weekOfRefineCards: CardProps[] = [
    {
        title: "Docker for Beginners",
        imgURL: "https://refine.ams3.cdn.digitaloceanspaces.com/week-of-refine/week-of-refine-invoicer.png",
        dateRange: "February 7, 2024",
        description:
            "Introduction to Docker and why we decided to use it for OpenPanel.",
        logo: StrapiWithText,
        bgLinearGradient:
            "bg-week-of-refine-strapi-card-light dark:bg-week-of-refine-strapi-card",
        link: "/week-of-refine-strapi",
    },
    {
        title: "Troubleshooting DNS",
        imgURL: "https://refine.ams3.cdn.digitaloceanspaces.com/week-of-refine/week-of-refine-pixels.png",
        dateRange: "December 14, 2023",
        description:
            "Troubleshooting DNS zone problems and solutions.",
        logo: SupabaseWithText,
        bgLinearGradient:
            "bg-week-of-refine-supabase-card-light dark:bg-week-of-refine-supabase-card",
        link: "/week-of-refine-supabase",
    },
];
