import React from 'react';
import AppContext from '../context/AppState'

export default class WelcomePanel extends React.Component {
    static contextType = AppContext

    constructor(props) {
        super(props);
        this.state = {
            key: ""
        };
    }

    componentDidMount() {
        window.addEventListener("message", this.receiveMessage)
    }

    receiveMessage = (event) => {
        const message = event.data.message;
        switch (message) {
            case 'jwtToken':
                console.log("Got the message!!", event.data.value)
                this.setState({key: event.data.value});
                this.context.set("jwtToken", event.data.value);
                break;
        }
    }

    render() {
        return <>
            <p>Jwt: {this.context.jwtToken}</p>
            <input value={this.state.key} readOnly></input>
        </>;
    }
}
