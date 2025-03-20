package php

const phpTpl = `<?php
/** Code generated from jsonrpc schema by rpcgen v{{ .Version }}; DO NOT EDIT. */

namespace {{.Namespace}};

	use EazyJsonRpc\BaseJsonRpcClient;
    use EazyJsonRpc\BaseJsonRpcException;
    use GuzzleHttp\Exception\GuzzleException;
    use JsonMapper_Exception;

	{{range .Classes}}
	/** {{.Name}} **/
	class {{.Name}} { {{range .Fields}}
	    /**{{if .Description}}
		* {{.Description}}{{end}}
	    * @var {{.Type}}{{if .Optional}}|null{{end}}
	    */
	    public ${{.Name}};{{end}}
	}
	{{end}}


    /**
     * RpcClient
     */
 	class RpcClient extends BaseJsonRpcClient {
		{{range .Methods}}
		/**
		* <{{.Name}}> RPC method{{range $row := .Description}}
		* {{$row}}{{end}}{{range .Parameters}}
		* @param {{.Type}}{{if .Optional}}|null{{end}} ${{.Name}}{{if .Optional}} [optional]{{end}}{{if .Description}} {{.Description}}{{end}}{{end}}
		* @param bool $isNotification [optional] set to true if call is notification
		* @return {{.Returns.Type}}{{if .Returns.Optional}}|null{{end}}
		* @throws BaseJsonRpcException
		* @throws GuzzleException
		* @throws JsonMapper_Exception
		**/
		public function {{.SafeName}}( {{range .Parameters}}{{if .Optional}}?{{end}}{{.ReturnType}} ${{.Name}}{{if .DefaultValue}} = {{.DefaultValue}}{{end}}, {{end}}$isNotification = false ){{if ne .Returns.Type "mixed"}}: {{if .Returns.Optional}}?{{end}}{{.Returns.ReturnType}}{{end}} {
			return $this->call( __FUNCTION__, {{if ne .Returns.BaseType .Returns.Type}} __NAMESPACE__ . '\\'. {{end}}'{{.Returns.Type}}', [ {{range .Parameters}}'{{.Name}}' => ${{.Name}}, {{end}}], $this->getRequestId( $isNotification ) );
		}
		{{end}}

        /**
         * Get Instance
         * @param $url string RPC server url
         * @return RpcClient
         */
        public static function GetInstance( $url ): RpcClient {
            return new self( $url );
        }
 	}
`
