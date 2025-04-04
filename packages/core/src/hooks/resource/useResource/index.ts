import { useContext } from "react";

import { ResourceContext } from "@contexts/resource";
import { useResourceWithRoute, useRouterContext } from "@hooks";

import type { BaseKey } from "../../../contexts/data/types";
import type { IResourceItem } from "../../../contexts/resource/types";
import type { ResourceRouterParams } from "../../../contexts/router/legacy/types";
import { useRouterType } from "../../../contexts/router/picker";
import type { Action } from "../../../contexts/router/types";
import { pickResource } from "../../../definitions/helpers/pick-resource";
import { useParsed } from "../../router/use-parsed";

export type UseResourceLegacyProps = {
  /**
   * Determines which resource to use for redirection
   * @deprecated resourceName deprecated. Use resourceNameOrRouteName instead # https://github.com/refinedev/refine/issues/1618
   */
  resourceName?: string;
  /**
   * Determines which resource to use for redirection
   * @default Resource name that it reads from route
   */
  resourceNameOrRouteName?: string;
  /**
   * Adds id to the end of the URL
   * @deprecated resourceName deprecated. Use resourceNameOrRouteName instead # https://github.com/refinedev/refine/issues/1618
   */
  recordItemId?: BaseKey;
};

/**
 * Matches the resource by identifier.
 * If not provided, the resource from the route will be returned.
 * If your resource does not explicitly define an identifier, the resource name will be used.
 */
export type UseResourceParam = string | undefined;

type SelectReturnType<T extends boolean> = T extends true
  ? { resource: IResourceItem; identifier: string }
  : { resource: IResourceItem; identifier: string } | undefined;

export type UseResourceReturnType = {
  resources: IResourceItem[];
  resource?: IResourceItem;
  /**
   * @deprecated Use `resource.name` instead when you need to get the resource name.
   */
  resourceName?: string;
  /**
   * @deprecated This value may not always reflect the correct "id" value. Use `useResourceParams` to obtain the calculated "id"` or `useParsed` to obtain the id from the route instead.
   */
  id?: BaseKey;
  /**
   * @deprecated This value may not always reflect the correct "action" value. Use `useResourceParams` to obtain the calculated "action" or `useParsed` to obtain the action from the route instead.
   */
  action?: Action;
  select: <T extends boolean = true>(
    resourceName: string,
    force?: T,
  ) => SelectReturnType<T>;
  identifier?: string;
};

type UseResourceReturnTypeWithResource = UseResourceReturnType & {
  resource: IResourceItem;
  identifier: string;
};

/**
 * @deprecated Use `useResource` with `identifier` property instead. (`identifier` does not check by route name in new router)
 */
export function useResource(
  props: UseResourceLegacyProps,
): UseResourceReturnType;
export function useResource(): UseResourceReturnType;
export function useResource<TIdentifier = UseResourceParam>(
  identifier: TIdentifier,
): TIdentifier extends NonNullable<UseResourceParam>
  ? UseResourceReturnTypeWithResource
  : UseResourceReturnType;
/**
 * `useResource` is used to get `resources` that are defined as property of the `<Refine>` component.
 *
 * @see {@link https://refine.dev/docs/api-reference/core/hooks/resource/useResource} for more details.
 */
export function useResource(
  args?: UseResourceLegacyProps | UseResourceParam,
): UseResourceReturnType {
  const { resources } = useContext(ResourceContext);
  const routerType = useRouterType();
  const params = useParsed();

  const selectResource = (resourceName: string, force = true) => {
    const isLegacy = routerType === "legacy";
    const pickedResource = pickResource(resourceName, resources, isLegacy);
    if (pickedResource) {
      return { resource: pickedResource, identifier: pickedResource.identifier ?? pickedResource.name };
    }
    if (force) {
      const fallbackResource = { name: resourceName, identifier: resourceName };
      return { resource: fallbackResource, identifier: fallbackResource.identifier };
    }
    return undefined;
  };

  // Legacy Router Logic
  if (routerType === "legacy") {
    const legacyResource = selectResource(params.resource ?? args?.resourceNameOrRouteName);
    return { ...legacyResource, resources, id: params.id, action: params.action };
  }

  // New Router Logic
  const identifier = typeof args === "string" ? args : args?.resourceNameOrRouteName;
  const resource = identifier ? selectResource(identifier)?.resource : undefined;

  return { resources, resource, id: params.id, action: params.action, identifier: resource?.identifier };
}
