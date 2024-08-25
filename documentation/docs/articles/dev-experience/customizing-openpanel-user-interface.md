# Customizing OpenPanel Interface (Branding and White-Label)

Everything in OpenPanel is modular and can easily be modified or disabled without breaking the rest of the functionalities.

To customize OpenPanel, you have the following options:

- Enable/disable features and pages from the OpenPanel interface.
- Set pre-installed services for user.
- Localize the interface.
- Set custom branding and logo.
- Set a custom color scheme.
- Replace How-to articles with your knowledge base.
- Customize any page.
- Customize login pages.
- Add custom CSS or JS code to the interface.
- Create custom functionality and pages in the interface.



## Enable/disable features

Administrators have the ability to enable or disable each feature (page) in the OpenPanel interface. To activate a feature, navigate to OpenAdmnin > Settings > OpenPanel](/docs/admin/settings/openpanel/#enable-features) and select service name in the "Enable Features" section and click save. 

Once enabled, the feature becomes instantly available to all users, appearing in the OpenPanel interface sidebar, search results, and dashboard icons.



## Set pre-installed services for user

OpenPanel uses [docker images](https://dev.openpanel.com/images/) as the base for each hosting plan. Based on the docker image, different services can be set per plan/user. For examples, we provide 2 docker iamges, one that has nginx pre-installed and another that uses apache. By creating a custom docker image, you can set in that image what to be pre-installed when you create a new user, for exmaple, set mariadb instead of mysql or install php ioncube loader extension.

To add a custom service pre-installed for users:

- [Create a custom docker image](/docs/articles/docker/building_a_docker_image_example_include_php_ioncubeloader/)
- [Create a new hosting plan with that docker image](/docs/admin/plans/hosting_plans/#create-a-plan)
- [Create a new user on that plan](/docs/admin/users/openpanel/#create-users)


## Localize the interface

## Set custom branding and logo

## Set a custom color scheme

## Replace How-to articles with your knowledge base

## Edit any page template

## Customize login page

## Add custom CSS or JS code

## Create custom pages
