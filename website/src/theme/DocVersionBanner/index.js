/**
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
import React from 'react';
import clsx from 'clsx';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Link from '@docusaurus/Link';
import Translate from '@docusaurus/Translate';
import {
  useActivePlugin,
  useDocVersionSuggestions,
} from '@docusaurus/plugin-content-docs/client';
import {ThemeClassNames} from '@docusaurus/theme-common';
import {
  useDocsPreferredVersion,
  useDocsVersion,
} from '@docusaurus/theme-common/internal';
function UnreleasedVersionLabel({siteTitle, versionMetadata}) {
  return (
    <Translate
      id="theme.docs.versions.unreleasedVersionLabel"
      description="The label used to tell the user that he's browsing an unreleased doc version"
      values={{
        siteTitle,
        versionLabel: <b>{versionMetadata.label}</b>,
      }}>
      {
        'This is unreleased documentation for {siteTitle} {versionLabel} version.'
      }
    </Translate>
  );
}
function MigratingVersionLabel({siteTitle, versionMetadata}) {
  return (
    <Translate
      id="theme.docs.versions.migratingVersionLabel"
      description="The label used to tell the user this doc version is still maintained but migration to a newer version is recommended"
      values={{
        siteTitle,
        versionLabel: <b>{versionMetadata.label}</b>,
        supportLink: (
          <Link to="/support">
            <Translate
              id="theme.docs.versions.migratingVersionSupportLinkLabel"
              description="The label used for the support link in the migrating version banner">
              reach out to support
            </Translate>
          </Link>
        ),
      }}>
      {
        'This is documentation for {siteTitle} {versionLabel}, which is still maintained. Switching to 2.0.0 is recommended, and we can help with migration — just {supportLink}.'
      }
    </Translate>
  );
}
function UnmaintainedVersionLabel({siteTitle, versionMetadata}) {
  return (
    <Translate
      id="theme.docs.versions.unmaintainedVersionLabel"
      description="The label used to tell the user that he's browsing an unmaintained doc version"
      values={{
        siteTitle,
        versionLabel: <b>{versionMetadata.label}</b>,
      }}>
      {
        'This is documentation for {siteTitle} {versionLabel}, which is no longer actively maintained.'
      }
    </Translate>
  );
}
const BannerLabelComponents = {
  unreleased: UnreleasedVersionLabel,
  unmaintained: UnmaintainedVersionLabel,
};
function BannerLabel(props) {
  if (props.versionMetadata.name === '1.X.X') {
    return <MigratingVersionLabel {...props} />;
  }
  const BannerLabelComponent =
    BannerLabelComponents[props.versionMetadata.banner];
  return <BannerLabelComponent {...props} />;
}
function LatestVersionSuggestionLabel({versionLabel, to, onClick}) {
  return (
    <Translate
      id="theme.docs.versions.latestVersionSuggestionLabel"
      description="The label used to tell the user to check the latest version"
      values={{
        versionLabel,
        latestVersionLink: (
          <b>
            <Link to={to} onClick={onClick}>
              <Translate
                id="theme.docs.versions.latestVersionLinkLabel"
                description="The label used for the latest version suggestion link label">
                latest version
              </Translate>
            </Link>
          </b>
        ),
      }}>
      {
        'For up-to-date documentation, see the {latestVersionLink} ({versionLabel}).'
      }
    </Translate>
  );
}
function DocVersionBannerEnabled({className, versionMetadata}) {
  const {
    siteConfig: {title: siteTitle},
  } = useDocusaurusContext();
  const {pluginId} = useActivePlugin({failfast: true});
  const getVersionMainDoc = (version) =>
    version.docs.find((doc) => doc.id === version.mainDocId);
  const {savePreferredVersionName} = useDocsPreferredVersion(pluginId);
  const {latestDocSuggestion, latestVersionSuggestion} =
    useDocVersionSuggestions(pluginId);
  // Try to link to same doc in latest version (not always possible), falling
  // back to main doc of latest version
  const latestVersionSuggestedDoc =
    latestDocSuggestion ?? getVersionMainDoc(latestVersionSuggestion);
  const isMigrating = versionMetadata.name === '1.X.X';
  return (
    <div
      className={clsx(
        className,
        ThemeClassNames.docs.docVersionBanner,
        `alert alert--${isMigrating ? 'info' : 'warning'} margin-bottom--md`,
      )}
      role="alert">
      <div>
        <BannerLabel siteTitle={siteTitle} versionMetadata={versionMetadata} />
      </div>
      {!isMigrating && (
        <div className="margin-top--md">
          <LatestVersionSuggestionLabel
            versionLabel={latestVersionSuggestion.label}
            to={latestVersionSuggestedDoc.path}
            onClick={() =>
              savePreferredVersionName(latestVersionSuggestion.name)
            }
          />
        </div>
      )}
    </div>
  );
}
export default function DocVersionBanner({className}) {
  const versionMetadata = useDocsVersion();
  if (versionMetadata.banner || versionMetadata.name === '1.X.X') {
    return (
      <DocVersionBannerEnabled
        className={className}
        versionMetadata={versionMetadata}
      />
    );
  }
  return null;
}
