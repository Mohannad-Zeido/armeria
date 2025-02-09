<template>
    <div class="object-editor" :class="{ 'visible': objectEditorOpen }">
        <div class="header">
            <div class="name">
                <span class="type">{{ objectEditorData.objectType }}</span>
                {{ objectEditorData.name }}
            </div>
            <div class="close" @click="handleClose">X</div>
        </div>
        <div class="properties">
            <div class="prop-container" v-for="prop in objectEditorData.properties" :key="prop.name">
                <div class="prop-name">{{ prop.name }}</div>
                <div class="prop-value">
                    <!-- editable type -->
                    <div
                        class="editable"
                        v-if="prop.propType == 'editable'"
                        @click="handleEditablePropClick($event, prop)"
                        @blur="handleEditablePropBlur($event)"
                        @keydown="handleEditablePropKeyDown($event, prop)"
                    >
                        {{ prop.value || "&nbsp;" }}
                    </div>
                    <!-- picture type -->
                    <div
                            class="picture"
                            ref="picture"
                            :style="{ backgroundImage: getBackgroundUrl(prop.value) }"
                            v-if="prop.propType == 'picture'"
                            @dragenter.stop.prevent="handlePictureDragEnter"
                            @drop.stop.prevent="handlePictureDragDrop"
                            @dragleave.stop.prevent="handlePictureDragLeave"
                            @dragover.stop.prevent
                    >
                    </div>
                    <!-- script type -->
                    <div
                            class="script"
                            v-if="prop.propType == 'script'"
                            @click="handleScriptEditClick"
                    >
                        Edit Script
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    import { mapState } from 'vuex';

    export default {
        name: 'ObjectEditor',
        computed: mapState(['isProduction', 'objectTarget', 'objectEditorOpen', 'objectEditorData']),
        data: function() {
            return {
                propOriginal: '',
            };
        },
        watch: {
            objectEditorOpen: function(newVal) {
                this.$socket.sendObj({
                    type: "objectEditorOpen",
                    payload: newVal
                })
            }
        },
        methods: {
            getBackgroundUrl(objectKey) {
                if (!this.isProduction) {
                    return `url(http://localhost:8081/oi/${objectKey})`;
                }

                return `url(/oi/${objectKey})`;
            },

            handleClose: function() {
                this.$store.dispatch('setObjectEditorOpen', false);
            },

            handleEditablePropClick: function(e, prop) {
                const editableDiv = e.target;
                editableDiv.contentEditable = 'true';
                editableDiv.focus();
                document.execCommand('selectAll', false, null);
                editableDiv.classList.add('editing');

                this.propOriginal = prop.value;
                this.$store.dispatch('setAllowGlobalHotkeys', false);
            },

            handleEditablePropBlur: function(e) {
                const editableDiv = e.target;
                editableDiv.contentEditable = 'false';
                editableDiv.classList.remove('editing');

                if (!editableDiv.classList.contains('success')) {
                    this.animateDivWithClass(e.target, 'failure');
                    editableDiv.innerHTML = this.propOriginal;
                }

                this.$store.dispatch('setAllowGlobalHotkeys', true);
            },

            handleEditablePropKeyDown: function(e, prop) {
                if (e.key === 'Enter') {
                    e.stopPropagation();
                    e.preventDefault();
                    this.animateDivWithClass(e.target, 'success');
                    e.target.blur();
                    this.setProperty(prop.name, e.target.innerHTML);
                } else if (e.key === 'Escape') {
                    e.stopPropagation();
                    e.preventDefault();
                    e.target.blur();
                }
            },

            animateDivWithClass: function(div, className) {
                div.classList.add(className);
                setTimeout(() => {
                    div.classList.add('anim');
                    div.classList.remove(className);
                }, 50);
                setTimeout(() => {
                    div.classList.remove('anim');
                }, 500);
            },

            setProperty: function(propName, propValue) {
                switch(this.objectEditorData.objectType) {
                    case 'room':
                        this.$socket.sendObj({
                            type: 'command',
                            payload: `/room set ${propName} ${propValue}`
                        });
                        break;
                    case 'character':
                        this.$socket.sendObj({
                            type: 'command',
                            payload: `/character set "${this.objectEditorData.name}" ${propName} ${propValue}`
                        });
                        break;
                    case 'mob':
                        this.$socket.sendObj({
                            type: 'command',
                            payload: `/mob set "${this.objectEditorData.name}" ${propName} ${propValue}`
                        });
                        break;
                    case 'item':
                        this.$socket.sendObj({
                            type: 'command',
                            payload: `/item set "${this.objectEditorData.name}" ${propName} ${propValue}`
                        });
                        break;
                }
            },

            handlePictureDragEnter: function(e) {
                e.target.classList.add("candrop");
            },

            handlePictureDragLeave: function(e) {
                e.target.classList.remove("candrop");
            },

            handlePictureDragDrop: function(e) {
                e.target.classList.remove("candrop");

                const files = event.dataTransfer.files;

                if (files.length > 1) {
                    this.$store.dispatch('showText', { data: '\nYou can only upload a single picture to this object.' });
                    return;
                }

                const file = files[0];
                const validTypes = ['image/jpeg', 'image/png'];

                if (validTypes.indexOf(file.type) === -1) {
                    this.$store.dispatch('showText', { data: '\nYou can only upload an image to this object.' });
                    return;
                }

                if (file.size > 512000) {
                    this.$store.dispatch('showText', { data: '\nYou cannot upload images that exceed 512KB.' });
                    return;
                }

                const reader = new FileReader()
                reader.readAsBinaryString(file)
                reader.onload = () => {
                    this.$socket.sendObj({
                        type: 'objectPictureUpload',
                        payload: {
                            objectType: this.objectEditorData.objectType,
                            name: this.objectEditorData.name,
                            pictureType: file.type,
                            pictureData: btoa(reader.result)
                        }
                    });
                }
            },

            handleScriptEditClick: function() {
                let baseUrl = '/scripteditor.html';
                let dev = 'false';

                if (!this.isProduction) {
                    baseUrl = 'http://localhost:8080/scripteditor.html';
                    dev = 'true';
                }

                window.open(
                    `${baseUrl}?name=${this.objectEditorData.name}&type=${this.objectEditorData.objectType}&accessKey=${this.objectEditorData.accessKey}&dev=${dev}`,
                    'scripteditor',
                    'width=800,height=600'
                );
            }
        }
    }
</script>

<style lang="scss" scoped>
    $hoverColor: #1f1e1e;
    $padding: 8px;

    .object-editor {
        background-color: #0b0b0b;
        flex-basis: 0px;
        overflow: hidden;
        transition: all .2s ease-in-out;
        box-shadow: 0px 0px 5px 0px #000;
    }

    .object-editor.visible {
        flex-basis: 300px;
        min-width: 300px;
        margin-right: 5px;
    }

    .header {
        border-bottom: 1px solid #313131;
        padding: 10px;
        background-color: #1c1c1c;
        display: flex;
    }

    .header .name {
        font-weight: 600;
        font-size: 16px;
        flex-grow: 1;
    }

    .header .name .type {
        font-weight: 400;
        font-size: 12px;
        text-transform: uppercase;
        background-color: #333;
        padding: 4px 7px;
        border-radius: 10px;
        margin-right: 5px;
    }

    .header .close {
        flex-shrink: 1;
        color: #ccc;
        transition: all .1s ease-in-out;
        font-weight: 600;
        background-color: #383838;
        padding: 0 7px;
        border-radius: 5px;
    }

    .header .close:hover {
        cursor: pointer;
        color: #fff;
        background-color: #5d5b5b;
    }

    .prop-container {
        display: flex;
        border-bottom: 1px solid #313131;
        border-right: 1px solid #313131;
        border-left: 1px solid #313131;
        font-size: 12px;
    }

    .prop-name {
        flex-basis: 100px;
        min-width: 100px;
        padding: 8px;
        background-color: #222;
    }

    .prop-value {
        flex-grow: 1;
        background-color: #111;
    }

    .prop-value .editable {
        padding: $padding;
        text-overflow: ellipsis;
        overflow: hidden;
        white-space: nowrap;
        max-width: 166px;
    }

    .prop-value .editable.anim {
        transition: all .5s ease-in-out;
    }

    .prop-value .editable.editing {
        text-overflow: clip;
        background-color: #000000 !important;
        box-shadow: 0px 0px 5px #737373;
    }

    .prop-value .editable.success {
        background-color: #317331 !important;
    }

    .prop-value .editable.failure {
        background-color: #7b2a2a !important;
    }

    .prop-value .editable:hover {
        cursor: pointer;
        color: #fff;
        background-color: $hoverColor;
    }

    .prop-value .editable:focus {
        outline: none;
    }

    .prop-value .picture {
        width: 75px;
        height: 75px;
        box-shadow: inset 0px 0px 5px 0px #3a3a3a;
        background-size: contain;
        padding: 8px;
    }

    .prop-value .picture.candrop {
        box-shadow: inset 0px 0px 5px 0px #ffe500;
    }

    .prop-value .script {
        padding: $padding;
        overflow: hidden;
        max-width: 166px;
    }

    .prop-value .script:hover {
        background-color: $hoverColor;
        cursor: pointer;
        color: #fff;
    }
</style>