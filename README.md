# matrix-on-call-bot
matrix-on-call-bot is a [matrix](https://matrix.org) bot for managing on-call stuff such as managing shifts and followups.

## Usage 
Just add the bot to a group and start using it by its commands.

## Commands
| command                                                           | description                                                                                               |
|-------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------|
| !help                                                             | show description of all commands                                                                          |
| !startshift [mentioned on calls]                                  | start a new shift with the mentioned people. If no one mentions the sender of the message will be on call |
| !listshifts                                                       | list all shifts                                                                                           |
| !endshift [shift id]                                              | end a shift                                                                                               |
| !followup [category: incoming/outgoing] [initiator] [description] | create a new follow up                                                                                    |
| !listfollowups                                                    | list all follow ups                                                                                       |
| !resolvefollowup [id]                                             | resolve a follow up                                                                                       |
| !report                                                           | Report this month shifts                                                                                  |