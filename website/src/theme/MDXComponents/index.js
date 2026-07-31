import MDXComponents from "@theme-original/MDXComponents";

import { BannerRandom } from "@site/src/components/banner/banner-random";
import DiscordBanner from "@site/src/components/blog/discord-banner";
import GithubBanner from "@site/src/components/blog/github-banner";
import TwitterBanner from "@site/src/components/blog/twitter-banner";
import GeneralConceptsLink from "@site/src/components/general-concepts-link";
import { GlobalConfigBadge } from "@site/src/components/global-config-badge";
import { GuideBadge } from "@site/src/components/guide-badge";
import PropTag from "@site/src/components/prop-tag";
import { RouterBadge } from "@site/src/components/router-badge";
import CommonDetails from "@site/src/refine-theme/common-details";
import CommonSummary from "@site/src/refine-theme/common-summary";
import CommonTabItem from "@site/src/refine-theme/common-tab-item";
import CommonTabs from "@site/src/refine-theme/common-tabs";
import { Blockquote } from "@site/src/refine-theme/common-blockquote";
import { Image } from "@site/src/components/image";
import { Table, FullTable } from "@site/src/refine-theme/common-table";
import { CreateRefineAppCommand } from "@site/src/partials/npm-scripts/create-refine-app-command.tsx";
import { InstallPackagesCommand } from "@site/src/partials/npm-scripts/install-packages-commands";

export default {
    ...MDXComponents,
    DiscordBanner: DiscordBanner,
    GithubBanner: GithubBanner,
    TwitterBanner: TwitterBanner,
    PropTag: PropTag,
    details: CommonDetails,
    summary: CommonSummary,
    Tabs: CommonTabs,
    TabItem: CommonTabItem,
    blockquote: Blockquote,
    GeneralConceptsLink,
    BannerRandom,
    GuideBadge,
    RouterBadge,
    GlobalConfigBadge,
    Image,
    table: Table,
    CreateRefineAppCommand: CreateRefineAppCommand,
    InstallPackagesCommand: InstallPackagesCommand,
    FullTable: FullTable,
};
