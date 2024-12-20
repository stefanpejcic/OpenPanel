import { gql } from "@urql/core";

import dataProvider from "../../src/index";
import client from "../gqlClient";
import "./deleteOne.mock";

const gqlMutation = gql`
  mutation DeleteOneBlogPost($input: DeleteOneBlogPostInput!) {
    deleteOneBlogPost(input: $input) {
      id
      title
    }
  }
`;

describe("deleteOne", () => {
  describe("with correct params", () => {
    it("works as expected", async () => {
      const { data } = await dataProvider(client).deleteOne({
        resource: "blogPosts",
        id: "42",
        meta: {
          gqlMutation,
        },
      });

      expect(data.id).toEqual(null);
      expect(data.title).toBeDefined();
    });
  });

  describe("without operation", () => {
    it("throws error", async () => {
      expect(
        dataProvider(client).deleteOne({ resource: "blogPosts", id: 42 }),
      ).rejects.toEqual(new Error("Operation is required."));
    });
  });
});
