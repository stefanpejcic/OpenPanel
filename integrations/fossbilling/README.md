# FOSSBilling-OpenPanel Server Manager

> [!NOTE]  
> Tested with [FOSSBilling](https://github.com/FOSSBilling/FOSSBilling) v0.6.22
> 



## Installation

- Download or git clone the OpenPanel.php file to your [FOSSBilling](https://github.com/FOSSBilling/FOSSBilling) installation at the following location: /library/Server/Manager


## Features
#### Server
- ✅ Verify Connection
- ✅ Create account
- ✅ Cancel account
- ✅ Suspend/Unsuspend account
- ✅ Change account package
- ✅ Change account password

#### Website Functions
- ❌ Create Website (This will also create the user in OpenPanel)
- ❌ Change Website Package
- ❌ Suspend/Un-suspend Website

#### User Functions
- ❌ Change User Password

#### Things The Don't Work Due To Lack Of API
- ❌ Changing Account Username
- ❌ Changing Account Domain
- ❌ Synchronizing Accounts

## Important Notes

- This community-maintained package isn't affiliated with [FOSSBilling](https://github.com/FOSSBilling/FOSSBilling). Please report issues here rather than on the [FOSSBilling](https://github.com/FOSSBilling/FOSSBilling) repo.
- Reseller support in the API is limited.  It does support creating Reseller accounts, though the API doesn't seem to provide a way to get all the domains/users hosted by the Reseller to suspend/un-suspend them. I'm not really sure if this happens if the reseller's website get suspended. 
- For questions, concerns, or issues with this server manager, please open an issue on GitHub.
- If [FOSSBilling](https://github.com/FOSSBilling/FOSSBilling) updates and breaks this server manager, report the issue here, and I'll update it to work with the latest version.
