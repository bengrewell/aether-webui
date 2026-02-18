# Controller Refactor

This is a major rewrite of much of the codebase to move towards more of a controller based architecture where there is
a central controller that manages the 'ServiceAPI' elements and the 'Provider' elements and wires them together on start
and manages their lifecycles. 

## Parts

### Provider

**Inputs**
- Logger
- Settings
- StateDB


**Outputs**
- Endpoints
- Initialization Error

**Functions**
- Name()
- Enable()
- Disable()
- Status()
- Start()
- Stop()

### ServiceAPI

**Inputs**
- Logger
- Settings
- Endpoints

**Outputs**
- Initialization Error

**Functions**
- Start()
- Stop()