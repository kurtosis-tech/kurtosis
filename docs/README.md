Kurtosis Docs
=============

>📖 Kurtosis documentation site codebase.
>
>👉 Read at [docs.kurtosis.com][kurtosis-docs]

---

### Philosophy
These docs strive to follow [the Diataxis framework](https://diataxis.fr/). In our docs, our categories behave like so:

- **Tutorials:**
  - Step-by-step walkthroughs (usually longer) for teaching the user that we intend users to complete once
  - Should take the user from "I don't even know what questions to ask" to "Now I have questions about the system"
  - Intended for the user to learn through doing; combination of a little "do this" and a little "this is what we did and how"
  - Expect users to come in with zero knowledge, and learn by doing
  - Focus more on the practical than long-form explanation (because that would be an Explanation)
- **Explanations:**
  - Long-form conveyance of knowledge
  - Should take the user from "I have questions" to "My questions are answered" or "I have more questions about other parts of the system" (they can then read other Explanations)
  - We generally expect the users to read once and internalize, because if the user needs to keep coming back to it then we explained it poorly
    - It's bad for the user to keep coming back to an Explanation because Explanations are very wordy and hard to search through; we would want them to keep coming back to Reference instead
  - Idea is to explain the reasoning and understanding behind the systems so that the user deeply internalizes it
  - Users don't know what 
  - Focuses more on the theoretical than the practical
- **Reference:** 
  - Easily-searchable, informative, in-case-you-forgot punchy nuggets of information where the user can get exactly the information they need
  - Should take the user from "I need quick information on X" to "I have that information"
  - **The user knows exactly what they need; we just need to give them the information**
  - Explicitly NOT step-by-step!
  - E.g.:
    - "What was the syntax of the `kurtosis.yml` again?"
    - API and syntax documentation
    - Bullet point information about how the packaging system works
- **Guides:** 
  - Step-by-step, "do this then this then this" instructions for how to complete workflows that users won't necessarily remember
  - Should take the user from "I need to know how to do X" to "I have done X"
  - User knows exactly the workflow they want to complete but not the steps to link it together (because if they did, they could use Reference)
  - NOT an Explanation; user is NOT looking to understand the deeper mechanics of the system (they either already know or don't care)

### Local Development
Install dependencies:
```shell
$ yarn
```

Validate and build the docs into the `build` directory (which can be served using any static content service):
```shell
$ yarn build
```

Start a local development server and open a browser window; most changes are reflected live without having to restart the server:
```shell
$ yarn start
```

Serve the `build` directory. This is useful to verifying the production build locally.
```shell
$ yarn serve
```

When your PR merges into `main`, the documentation will automatically be rebuilt and republished to [our docs page][kurtosis-docs].

<!------ ONLY LINKS BELOW HERE ------------>
[kurtosis-docs]: https://docs.kurtosis.com
